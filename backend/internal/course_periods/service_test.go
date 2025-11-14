package course_periods

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/course_periods/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewCoursePeriodsService(t *testing.T) {
	type testCase struct {
		name       string
		repository Repository
		logger     *zap.Logger
	}
	tc := testCase{
		name:       "test",
		repository: mocks.NewMockRepository(t),
		logger:     zap.NewNop(),
	}
	t.Run(tc.name, func(t *testing.T) {
		got := NewCoursePeriodsService(tc.repository, tc.logger)
		assert.NotNil(t, got)
		assert.Equal(t, tc.repository, got.repo)
		assert.Equal(t, tc.logger, got.log)
	})
}

func TestCoursePeriodService_GetCoursePeriod(t *testing.T) {
	type testCase struct {
		name     string
		periodID string
		prepare  func(ctx context.Context, repoMock *mocks.MockRepository, periodID string)
		want     entities.CoursePeriod
		wantErr  error
	}

	tests := []testCase{
		{
			name:     "success",
			periodID: "period-id",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, periodID string) {
				repoMock.EXPECT().GetCoursePeriodByID(ctx, periodID).RunAndReturn(
					func(ctx context.Context, id string) (entities.CoursePeriod, error) {
						return entities.CoursePeriod{ID: "period-id"}, nil
					})
				repoMock.EXPECT().GetAnnouncementsByCoursePeriodID(ctx, periodID).RunAndReturn(
					func(ctx context.Context, id string) ([]entities.Announcement, error) {
						return []entities.Announcement{{ID: "announcement-id"}}, nil
					})
			},
			want: entities.CoursePeriod{
				ID: "period-id",
				Announcements: []entities.Announcement{
					{
						ID: "announcement-id",
					},
				},
			},
			wantErr: nil,
		},
		{
			name:     "error getting course period",
			periodID: "period-id",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, periodID string) {
				repoMock.EXPECT().GetCoursePeriodByID(ctx, periodID).RunAndReturn(
					func(ctx context.Context, id string) (entities.CoursePeriod, error) {
						return entities.CoursePeriod{}, errors.New("error getting course period")
					})
			},
			want:    entities.CoursePeriod{},
			wantErr: errors.New("error getting course period"),
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock, tt.periodID)
			}
			s := NewCoursePeriodsService(repoMock, loggerMock)
			got, err := s.GetCoursePeriod(ctx, tt.periodID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoursePeriodService_GetCoursePeriods(t *testing.T) {
	type testCase struct {
		name      string
		periodID  string
		prepare   func(ctx context.Context, repoMock *mocks.MockRepository, periodID string, pageScope entities.PageScope)
		want      []entities.CoursePeriod
		pageScope entities.PageScope
		wantErr   error
	}

	tests := []testCase{
		{
			name:     "success",
			periodID: "period-id",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, periodID string, pageScope entities.PageScope) {
				repoMock.EXPECT().GetCoursePeriods(ctx, periodID, pageScope).RunAndReturn(
					func(ctx context.Context, courseID string, ps entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
						return []entities.CoursePeriod{{ID: periodID}}, pageScope, nil
					})
				repoMock.EXPECT().GetAnnouncementsByCoursePeriodID(ctx, periodID).RunAndReturn(
					func(ctx context.Context, id string) ([]entities.Announcement, error) {
						return []entities.Announcement{{ID: "announcement-id"}}, nil
					})
			},
			pageScope: entities.PageScope{},
			want: []entities.CoursePeriod{
				{
					ID: "period-id",
					Announcements: []entities.Announcement{
						{
							ID: "announcement-id",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:     "error getting course period",
			periodID: "period-id",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, periodID string, pageScope entities.PageScope) {
				repoMock.EXPECT().GetCoursePeriods(ctx, periodID, pageScope).RunAndReturn(
					func(ctx context.Context, courseID string, ps entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
						return nil, entities.PageScope{}, errors.New("error getting course period")
					})
			},
			wantErr: errors.New("error getting course period"),
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock, tt.periodID, tt.pageScope)
			}
			s := NewCoursePeriodsService(repoMock, loggerMock)
			got, _, err := s.GetCoursePeriods(ctx, tt.periodID, tt.pageScope)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoursePeriodService_CreateCoursePeriod(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(ctx context.Context, repoMock *mocks.MockRepository)
		want    int64
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().CreateCoursePeriod(ctx, mock.AnythingOfType("entities.CoursePeriod")).RunAndReturn(
					func(ctx context.Context, period entities.CoursePeriod) (int64, error) {
						return int64(2), nil
					})
			},
			want:    int64(2),
			wantErr: nil,
		},
		{
			name: "error creating new course period",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().CreateCoursePeriod(ctx, mock.AnythingOfType("entities.CoursePeriod")).RunAndReturn(
					func(ctx context.Context, period entities.CoursePeriod) (int64, error) {
						return int64(-1), errors.New("error creating new course period")
					})
			},
			want:    int64(-1),
			wantErr: errors.New("error creating new course period"),
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock)
			}
			s := NewCoursePeriodsService(repoMock, loggerMock)
			got, err := s.CreateCoursePeriod(ctx, entities.CoursePeriod{})
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoursePeriodService_UpdateCoursePeriod(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(ctx context.Context, repoMock *mocks.MockRepository)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().UpdateCoursePeriod(ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).RunAndReturn(
					func(ctx context.Context, periodID string, period entities.CoursePeriod) error {
						return nil
					})
			},
			wantErr: nil,
		},
		{
			name: "error updating course period",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().UpdateCoursePeriod(ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).RunAndReturn(
					func(ctx context.Context, periodID string, period entities.CoursePeriod) error {
						return errors.New("error updating course period")
					})
			},
			wantErr: errors.New("error updating course period"),
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock)
			}
			s := NewCoursePeriodsService(repoMock, loggerMock)
			err := s.UpdateCoursePeriod(ctx, "1", entities.CoursePeriod{})
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCoursePeriodService_DeleteCoursePeriod(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(ctx context.Context, repoMock *mocks.MockRepository)
		wantErr error
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().DeleteCoursePeriod(ctx, mock.AnythingOfType("string")).RunAndReturn(
					func(ctx context.Context, periodID string) error {
						return nil
					})
			},
			wantErr: nil,
		},
		{
			name: "error deleting course period",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository) {
				repoMock.EXPECT().DeleteCoursePeriod(ctx, mock.AnythingOfType("string")).RunAndReturn(
					func(ctx context.Context, periodID string) error {
						return errors.New("error deleting course period")
					})
			},
			wantErr: errors.New("error deleting course period"),
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock)
			}
			s := NewCoursePeriodsService(repoMock, loggerMock)
			err := s.DeleteCoursePeriod(ctx, "1", "1")
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
