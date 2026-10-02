package jobs

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/jobs/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const testKind = "test_kind"

func newJob(attempts int) *entities.Job {
	return &entities.Job{ID: 7, Kind: testKind, Payload: []byte(`{"x":1}`), Attempts: attempts, MaxAttempts: 5}
}

func TestWorker_RunNext(t *testing.T) {
	tests := []struct {
		name    string
		handler Handler
		prepare func(repo *mocks.MockRepository)
		want    bool
	}{
		{
			name: "no job to run",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(nil, nil)
			},
			want: false,
		},
		{
			name: "claim error",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(nil, errors.New("db down"))
			},
			want: false,
		},
		{
			name: "success marks the job done and passes the payload",
			handler: func(_ context.Context, payload []byte) error {
				if string(payload) != `{"x":1}` {
					return errors.New("unexpected payload")
				}
				return nil
			},
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(1), nil)
				repo.EXPECT().CompleteJob(mock.Anything, int64(7)).Return(nil)
			},
			want: true,
		},
		{
			name:    "failure with attempts left retries with backoff",
			handler: func(context.Context, []byte) error { return errors.New("gotenberg down") },
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(2), nil)
				repo.EXPECT().RetryJobLater(mock.Anything, int64(7), "gotenberg down", 2*time.Minute).Return(nil)
			},
			want: true,
		},
		{
			name:    "failure on the last attempt fails the job",
			handler: func(context.Context, []byte) error { return errors.New("gotenberg down") },
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(5), nil)
				repo.EXPECT().FailJob(mock.Anything, int64(7), "gotenberg down").Return(nil)
			},
			want: true,
		},
		{
			name:    "panic is recorded as a failed attempt",
			handler: func(context.Context, []byte) error { panic("boom") },
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(1), nil)
				repo.EXPECT().RetryJobLater(mock.Anything, int64(7), "job panicked: boom", time.Minute).Return(nil)
			},
			want: true,
		},
		{
			name: "unknown kind fails the job without retrying",
			prepare: func(repo *mocks.MockRepository) {
				job := newJob(1)
				job.Kind = "other_kind"
				repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(job, nil)
				repo.EXPECT().FailJob(mock.Anything, int64(7), `no handler registered for job kind "other_kind"`).Return(nil)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			tt.prepare(repo)
			w := NewWorker(repo, zap.NewNop())
			if tt.handler != nil {
				w.Register(testKind, tt.handler)
			}

			assert.Equal(t, tt.want, w.RunNext(context.Background()))
		})
	}
}

// TestWorker_RunNext_Shutdown checks that a job interrupted by shutdown is left running, so its
// lease expires and it is claimed again on the next start.
func TestWorker_RunNext_Shutdown(t *testing.T) {
	repo := mocks.NewMockRepository(t)
	repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(1), nil)

	ctx, cancel := context.WithCancel(context.Background())
	w := NewWorker(repo, zap.NewNop())
	w.Register(testKind, func(ctx context.Context, _ []byte) error {
		cancel()
		return ctx.Err()
	})

	assert.True(t, w.RunNext(ctx))
	// The mock fails the test on any unexpected CompleteJob/RetryJobLater/FailJob call.
}

func TestWorker_Run_DrainsQueueUntilCancelled(t *testing.T) {
	repo := mocks.NewMockRepository(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runs int
	repo.EXPECT().ClaimJob(mock.Anything, defaultLease).Return(newJob(1), nil).Times(2)
	repo.EXPECT().CompleteJob(mock.Anything, int64(7)).Return(nil).Times(2)
	repo.EXPECT().ClaimJob(mock.Anything, defaultLease).RunAndReturn(func(context.Context, time.Duration) (*entities.Job, error) {
		cancel()
		return nil, nil
	}).Once()

	w := NewWorker(repo, zap.NewNop())
	w.Register(testKind, func(context.Context, []byte) error {
		runs++
		return nil
	})

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
	assert.Equal(t, 2, runs)
}

func TestRetryDelay(t *testing.T) {
	assert.Equal(t, time.Minute, retryDelay(1))
	assert.Equal(t, 2*time.Minute, retryDelay(2))
	assert.Equal(t, 8*time.Minute, retryDelay(4))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "short", truncate("short"))

	long := strings.Repeat("a", maxLastErrorLen-1) + "ñ" + "tail"
	got := truncate(long)
	assert.LessOrEqual(t, len(got), maxLastErrorLen)
	assert.True(t, strings.HasPrefix(long, got))
	assert.NotContains(t, got, "�")
}
