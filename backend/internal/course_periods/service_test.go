package course_periods

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/eaguilar88/deu/internal/course_periods/mocks"
// 	"github.com/eaguilar88/deu/internal/entities"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"go.uber.org/zap"
// )

// func TestNewCoursePeriodsService(t *testing.T) {
// 	type testCase struct {
// 		name       string
// 		repository Repository
// 		logger     *zap.Logger
// 		want       *CoursePeriodService
// 	}
// 	tc := testCase{
// 		name:       "test",
// 		repository: &mocks.MockRepository{},
// 		logger:     zap.NewNop(),
// 		want: &CoursePeriodService{
// 			repo: &mocks.MockRepository{},
// 			log:  zap.NewNop(),
// 		},
// 	}
// 	t.Run(tc.name, func(t *testing.T) {
// 		got := NewCoursePeriodsService(tc.repository, tc.logger)
// 		assert.Equal(t, tc.want, got)
// 	})
// }

// func TestCoursePeriodService_GetCoursePeriod(t *testing.T) {
// 	type testCase struct {
// 		name     string
// 		periodID string
// 		repoMock *mocks.MockRepository
// 		prepare  func(ctx context.Context, tc *testCase)
// 		want     entities.CoursePeriod
// 		wantErr  error
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "success",
// 			periodID: "period-id",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("GetCoursePeriodByID", ctx, tc.periodID).Return(entities.CoursePeriod{
// 					ID: "period-id",
// 				}, nil)
// 				tc.repoMock.On("GetUsersByCoursePeriodID", ctx, tc.periodID).Return([]entities.User{
// 					{
// 						ID: "user-id",
// 					},
// 				}, nil)
// 				tc.repoMock.On("GetAnnouncementsByCoursePeriodID", ctx, tc.periodID).Return([]entities.Announcement{
// 					{
// 						ID: "announcement-id",
// 					},
// 				}, nil)
// 			},
// 			want: entities.CoursePeriod{
// 				ID: "period-id",
// 				Participants: []entities.User{
// 					{
// 						ID: "user-id",
// 					},
// 				},
// 				Announcements: []entities.Announcement{
// 					{
// 						ID: "announcement-id",
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "error getting course period",
// 			periodID: "period-id",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("GetCoursePeriodByID", ctx, tc.periodID).Return(entities.CoursePeriod{}, errors.New("error getting course period"))
// 			},
// 			want:    entities.CoursePeriod{},
// 			wantErr: errors.New("error getting course period"),
// 		},
// 	}
// 	ctx := context.Background()
// 	loggerMock := zap.NewNop()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.prepare != nil {
// 				tt.prepare(ctx, &tt)
// 			}
// 			s := NewCoursePeriodsService(tt.repoMock, loggerMock)
// 			got, err := s.GetCoursePeriod(ctx, tt.periodID)
// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want, got)
// 			tt.repoMock.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCoursePeriodService_GetCoursePeriods(t *testing.T) {
// 	type testCase struct {
// 		name      string
// 		periodID  string
// 		repoMock  *mocks.MockRepository
// 		prepare   func(ctx context.Context, tc *testCase)
// 		want      []entities.CoursePeriod
// 		pageScope entities.PageScope
// 		wantErr   error
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "success",
// 			periodID: "period-id",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("GetCoursePeriods", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
// 					Return([]entities.CoursePeriod{{ID: tc.periodID}}, tc.pageScope, nil)
// 				tc.repoMock.On("GetUsersByCoursePeriodID", ctx, tc.periodID).
// 					Return([]entities.User{
// 						{
// 							ID: "user-id",
// 						},
// 					}, nil)
// 				tc.repoMock.On("GetAnnouncementsByCoursePeriodID", ctx, tc.periodID).
// 					Return([]entities.Announcement{
// 						{
// 							ID: "announcement-id",
// 						},
// 					}, nil)
// 			},
// 			pageScope: entities.PageScope{},
// 			want: []entities.CoursePeriod{
// 				{
// 					ID: "period-id",
// 					Participants: []entities.User{
// 						{
// 							ID: "user-id",
// 						},
// 					},
// 					Announcements: []entities.Announcement{
// 						{
// 							ID: "announcement-id",
// 						},
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "error getting period participants",
// 			periodID: "period-id",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("GetCoursePeriods", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
// 					Return([]entities.CoursePeriod{{ID: tc.periodID}}, tc.pageScope, nil)
// 				tc.repoMock.On("GetUsersByCoursePeriodID", ctx, tc.periodID).
// 					Return(nil, errors.New("error getting course period participants"))
// 			},
// 			pageScope: entities.PageScope{},
// 			wantErr:   errors.New("error getting course period participants"),
// 		},
// 		{
// 			name:     "error getting course period",
// 			periodID: "period-id",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("GetCoursePeriods", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.PageScope")).
// 					Return(nil, entities.PageScope{}, errors.New("error getting course period"))
// 			},
// 			wantErr: errors.New("error getting course period"),
// 		},
// 	}
// 	ctx := context.Background()
// 	loggerMock := zap.NewNop()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.prepare != nil {
// 				tt.prepare(ctx, &tt)
// 			}
// 			s := NewCoursePeriodsService(tt.repoMock, loggerMock)
// 			got, _, err := s.GetCoursePeriods(ctx, tt.periodID, tt.pageScope)
// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want, got)
// 			tt.repoMock.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCoursePeriodService_CreateCoursePeriod(t *testing.T) {
// 	type testCase struct {
// 		name     string
// 		repoMock *mocks.MockRepository
// 		prepare  func(ctx context.Context, tc *testCase)
// 		want     int64
// 		wantErr  error
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "success",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("CreateCoursePeriod", ctx, mock.AnythingOfType("entities.CoursePeriod")).
// 					Return(int64(2), nil)
// 			},
// 			want:    int64(2),
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "error creating new course period",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("CreateCoursePeriod", ctx, mock.AnythingOfType("entities.CoursePeriod")).
// 					Return(int64(-1), errors.New("error creating new course period"))
// 			},
// 			want:    int64(-1),
// 			wantErr: errors.New("error creating new course period"),
// 		},
// 	}
// 	ctx := context.Background()
// 	loggerMock := zap.NewNop()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.prepare != nil {
// 				tt.prepare(ctx, &tt)
// 			}
// 			s := NewCoursePeriodsService(tt.repoMock, loggerMock)
// 			got, err := s.CreateCoursePeriod(ctx, entities.CoursePeriod{})
// 			assert.Equal(t, tt.wantErr, err)
// 			assert.Equal(t, tt.want, got)
// 			tt.repoMock.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCoursePeriodService_UpdateCoursePeriod(t *testing.T) {
// 	type testCase struct {
// 		name     string
// 		repoMock *mocks.MockRepository
// 		prepare  func(ctx context.Context, tc *testCase)
// 		wantErr  error
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "success",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("UpdateCoursePeriod", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
// 					Return(nil)
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "error updating course period",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("UpdateCoursePeriod", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("entities.CoursePeriod")).
// 					Return(errors.New("error updating course period"))
// 			},
// 			wantErr: errors.New("error updating course period"),
// 		},
// 	}
// 	ctx := context.Background()
// 	loggerMock := zap.NewNop()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.prepare != nil {
// 				tt.prepare(ctx, &tt)
// 			}
// 			s := NewCoursePeriodsService(tt.repoMock, loggerMock)
// 			err := s.UpdateCoursePeriod(ctx, "1", entities.CoursePeriod{})
// 			assert.Equal(t, tt.wantErr, err)
// 			tt.repoMock.AssertExpectations(t)
// 		})
// 	}
// }

// func TestCoursePeriodService_DeleteCoursePeriod(t *testing.T) {
// 	type testCase struct {
// 		name     string
// 		repoMock *mocks.MockRepository
// 		prepare  func(ctx context.Context, tc *testCase)
// 		wantErr  error
// 	}

// 	tests := []testCase{
// 		{
// 			name:     "success",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("DeleteCoursePeriod", ctx, mock.AnythingOfType("string")).
// 					Return(nil)
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "error deleting course period",
// 			repoMock: &mocks.MockRepository{},
// 			prepare: func(ctx context.Context, tc *testCase) {
// 				tc.repoMock.On("DeleteCoursePeriod", ctx, mock.AnythingOfType("string")).
// 					Return(errors.New("error deleting course period"))
// 			},
// 			wantErr: errors.New("error deleting course period"),
// 		},
// 	}
// 	ctx := context.Background()
// 	loggerMock := zap.NewNop()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.prepare != nil {
// 				tt.prepare(ctx, &tt)
// 			}
// 			s := NewCoursePeriodsService(tt.repoMock, loggerMock)
// 			err := s.DeleteCoursePeriod(ctx, "1", "1")
// 			assert.Equal(t, tt.wantErr, err)
// 			tt.repoMock.AssertExpectations(t)
// 		})
// 	}
// }
