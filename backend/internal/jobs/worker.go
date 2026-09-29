package jobs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

const (
	defaultPollInterval = 3 * time.Second
	// defaultLease bounds how long a job may run. A job whose lease expires (e.g. the process
	// died mid-run) is claimed again, so handlers must be idempotent.
	defaultLease = 30 * time.Minute
	// baseRetryDelay is doubled on every attempt: 1m, 2m, 4m, 8m...
	baseRetryDelay = 30 * time.Second
	// maxLastErrorLen keeps an error's text from bloating the jobs table.
	maxLastErrorLen = 2000
)

// Repository is the job queue storage.
type Repository interface {
	ClaimJob(ctx context.Context, lease time.Duration) (*entities.Job, error)
	CompleteJob(ctx context.Context, id int64) error
	RetryJobLater(ctx context.Context, id int64, lastError string, delay time.Duration) error
	FailJob(ctx context.Context, id int64, lastError string) error
}

// Handler runs a job of one kind, given its payload.
type Handler func(ctx context.Context, payload []byte) error

// Worker polls the job queue and runs claimed jobs one at a time.
type Worker struct {
	repo         Repository
	handlers     map[string]Handler
	pollInterval time.Duration
	lease        time.Duration
	logger       *zap.Logger
}

func NewWorker(repo Repository, logger *zap.Logger) *Worker {
	return &Worker{
		repo:         repo,
		handlers:     make(map[string]Handler),
		pollInterval: defaultPollInterval,
		lease:        defaultLease,
		logger:       logger,
	}
}

// Register sets the handler for jobs of the given kind.
func (w *Worker) Register(kind string, handler Handler) {
	w.handlers[kind] = handler
}

// Run processes jobs until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	w.logger.Info("jobs worker started", zap.Duration("poll_interval", w.pollInterval))
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		// Drain every runnable job before waiting for the next tick.
		for ctx.Err() == nil {
			if !w.RunNext(ctx) {
				break
			}
		}

		select {
		case <-ctx.Done():
			w.logger.Info("jobs worker stopped")
			return
		case <-ticker.C:
		}
	}
}

// RunNext claims and runs one job. It reports whether a job was claimed.
func (w *Worker) RunNext(ctx context.Context) bool {
	job, err := w.repo.ClaimJob(ctx, w.lease)
	if err != nil {
		if ctx.Err() == nil {
			w.logger.Error("failed to claim job", zap.Error(err))
		}
		return false
	}
	if job == nil {
		return false
	}

	logger := w.logger.With(zap.Int64("job_id", job.ID), zap.String("kind", job.Kind), zap.Int("attempt", job.Attempts))
	logger.Info("running job")

	runErr := w.run(ctx, job)

	// The worker is shutting down: leave the job running so its lease expires and it is claimed
	// again on the next start.
	if ctx.Err() != nil {
		logger.Warn("job interrupted by shutdown", zap.Error(runErr))
		return true
	}

	// Record the outcome even if the job's own context timed out.
	storeCtx := context.WithoutCancel(ctx)
	switch {
	case runErr == nil:
		if err := w.repo.CompleteJob(storeCtx, job.ID); err != nil {
			logger.Error("failed to mark job as done", zap.Error(err))
		} else {
			logger.Info("job done")
		}
	case job.Attempts >= job.MaxAttempts || errors.Is(runErr, errUnknownKind):
		logger.Error("job failed permanently", zap.Error(runErr))
		if err := w.repo.FailJob(storeCtx, job.ID, truncate(runErr.Error())); err != nil {
			logger.Error("failed to mark job as failed", zap.Error(err))
		}
	default:
		delay := retryDelay(job.Attempts)
		logger.Warn("job failed, retrying later", zap.Error(runErr), zap.Duration("retry_in", delay))
		if err := w.repo.RetryJobLater(storeCtx, job.ID, truncate(runErr.Error()), delay); err != nil {
			logger.Error("failed to reschedule job", zap.Error(err))
		}
	}
	return true
}

var errUnknownKind = errors.New("no handler registered for job kind")

// run executes the job's handler within its lease, turning a panic into an error.
func (w *Worker) run(ctx context.Context, job *entities.Job) (err error) {
	handler, ok := w.handlers[job.Kind]
	if !ok {
		return fmt.Errorf("%w %q", errUnknownKind, job.Kind)
	}

	jobCtx, cancel := context.WithTimeout(ctx, w.lease)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job panicked: %v", r)
		}
	}()
	return handler(jobCtx, job.Payload)
}

// retryDelay is the wait before the next attempt: baseRetryDelay * 2^attempts.
func retryDelay(attempts int) time.Duration {
	return baseRetryDelay * time.Duration(math.Pow(2, float64(attempts)))
}

func truncate(s string) string {
	if len(s) <= maxLastErrorLen {
		return s
	}
	// Cutting may split a multi-byte rune, which Postgres would reject as invalid UTF-8.
	return strings.ToValidUTF8(s[:maxLastErrorLen], "")
}
