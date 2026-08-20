package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/activities/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestService(t *testing.T) (*service, *mocks.MockRepository, *mocks.MockStorageClient) {
	t.Helper()
	repo := mocks.NewMockRepository(t)
	storage := mocks.NewMockStorageClient(t)
	logger := zap.NewNop()
	svc := NewService(repo, storage, logger).(*service)
	return svc, repo, storage
}

func TestNewService(t *testing.T) {
	repo := mocks.NewMockRepository(t)
	storage := mocks.NewMockStorageClient(t)
	logger := zap.NewNop()

	got := NewService(repo, storage, logger)
	assert.NotNil(t, got)
	s := got.(*service)
	assert.Equal(t, repo, s.repo)
	assert.Equal(t, storage, s.storage)
	assert.Equal(t, logger, s.logger)
}

func TestService_GetActivity(t *testing.T) {
	type testCase struct {
		name         string
		id           string
		prepare      func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		want         entities.Activity
		wantCoverURL string
		wantErr      bool
		errTarget    error
	}

	tests := []testCase{
		{
			name: "success no cover image",
			id:   "1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{ID: "1", Name: "Test"}, nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeActivity).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    entities.Activity{ID: "1", Name: "Test"},
			wantErr: false,
		},
		{
			name: "success with cover image",
			id:   "1",
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				f := &entities.File{Key: "files/activities/1/cubierta.jpg", Name: "cubierta.jpg", Purpose: entities.ActivityFileTypeCoverImage}
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{ID: "1", Name: "Test"}, nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeActivity).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ActivityFileTypeCoverImage: {f}}, nil
					})
				storage.EXPECT().GetFileURL(mock.Anything, f.Key).
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "http://cdn/cubierta.jpg", nil
					})
			},
			wantCoverURL: "http://cdn/cubierta.jpg",
			wantErr:      false,
		},
		{
			name: "not found",
			id:   "99",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetActivityByID(mock.Anything, "99").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{}, ErrActivityNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrActivityNotFound,
		},
		{
			name: "db error",
			id:   "1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name: "get files error",
			id:   "1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{ID: "1"}, nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeActivity).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return nil, errors.New("storage error")
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, storage := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo, storage)
			}
			got, err := svc.GetActivity(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
				if tt.want.ID != "" {
					assert.Equal(t, tt.want.ID, got.ID)
				}
				if tt.wantCoverURL != "" {
					require.NotNil(t, got.CoverImage)
					assert.Equal(t, tt.wantCoverURL, got.CoverImage.URL)
				}
			}
		})
	}
}

func TestService_GetActivities(t *testing.T) {
	type testCase struct {
		name      string
		filter    entities.ActivityFilter
		pageScope entities.PageScope
		prepare   func(repo *mocks.MockRepository)
		wantLen   int
		wantErr   bool
	}

	tests := []testCase{
		{
			name:      "success with results",
			filter:    entities.ActivityFilter{GroupID: "1"},
			pageScope: entities.PageScope{Page: 1, PerPage: 10},
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivities(mock.Anything, entities.ActivityFilter{GroupID: "1"}, entities.PageScope{Page: 1, PerPage: 10}).
					RunAndReturn(func(_ context.Context, _ entities.ActivityFilter, ps entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
						return []entities.Activity{{ID: "1"}, {ID: "2"}}, ps, nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, mock.Anything, entities.OwnerTypeActivity).
					Return(entities.GroupedFiles{}, nil)
				repo.EXPECT().GetActivityMetrics(mock.Anything, "1").
					Return(entities.ActivityMetrics{}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:      "success empty result",
			filter:    entities.ActivityFilter{GroupID: "42"},
			pageScope: entities.PageScope{Page: 1, PerPage: 10},
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivities(mock.Anything, entities.ActivityFilter{GroupID: "42"}, entities.PageScope{Page: 1, PerPage: 10}).
					RunAndReturn(func(_ context.Context, _ entities.ActivityFilter, ps entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
						return nil, ps, nil
					})
				repo.EXPECT().GetActivityMetrics(mock.Anything, "42").
					Return(entities.ActivityMetrics{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:      "db error",
			filter:    entities.ActivityFilter{GroupID: "1"},
			pageScope: entities.PageScope{Page: 1, PerPage: 10},
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivities(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.ActivityFilter, _ entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
						return nil, entities.PageScope{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo)
			}
			got, _, _, err := svc.GetActivities(context.Background(), tt.filter, tt.pageScope)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestService_CreateActivity(t *testing.T) {
	type testCase struct {
		name     string
		activity entities.Activity
		prepare  func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		wantID   int64
		wantErr  bool
	}

	tests := []testCase{
		{
			name:     "success no cover image",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, entities.Activity{GroupID: "1", Name: "Workshop"}).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return 42, nil
					})
			},
			wantID:  42,
			wantErr: false,
		},
		{
			name: "success with cover image",
			activity: entities.Activity{
				GroupID:    "1",
				Name:       "Workshop",
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return 10, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, files []*entities.File) error {
						assert.Equal(t, "files/activities/10/cubierta.jpg", files[0].Key)
						assert.Equal(t, entities.OwnerTypeActivity, files[0].OwnerType)
						assert.Equal(t, entities.ActivityFileTypeCoverImage, files[0].Purpose)
						return nil
					})
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return nil
					})
			},
			wantID:  10,
			wantErr: false,
		},
		{
			name:     "repo create error",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return -1, errors.New("db error")
					})
			},
			wantID:  -1,
			wantErr: true,
		},
		{
			name: "storage upload error",
			activity: entities.Activity{
				GroupID:    "1",
				Name:       "Workshop",
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return 5, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("storage error")
					})
			},
			wantID:  -1,
			wantErr: true,
		},
		{
			name: "save cover image to DB error",
			activity: entities.Activity{
				GroupID:    "1",
				Name:       "Workshop",
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return 7, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return nil
					})
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("db error")
					})
			},
			wantID:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, storage := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo, storage)
			}
			got, err := svc.CreateActivity(context.Background(), tt.activity)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantID, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, got)
			}
		})
	}
}

func TestService_UpdateActivity(t *testing.T) {
	type testCase struct {
		name      string
		id        string
		activity  entities.Activity
		prepare   func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name:     "success",
			id:       "1",
			activity: entities.Activity{Name: "Updated Workshop"},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, entities.Activity{ID: "1", Name: "Updated Workshop"}).
					RunAndReturn(func(_ context.Context, _ entities.Activity) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:     "not found",
			id:       "99",
			activity: entities.Activity{Name: "Updated"},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) error {
						return ErrActivityNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrActivityNotFound,
		},
		{
			name:     "db error",
			id:       "1",
			activity: entities.Activity{},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name: "updates cover image only",
			id:   "1",
			activity: entities.Activity{
				Name:       "Updated Workshop",
				CreatedBy:  "user1",
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).Return(nil)
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, files []*entities.File) error {
						require.Len(t, files, 1)
						assert.Equal(t, "files/activities/1/cubierta.jpg", files[0].Key)
						assert.Equal(t, entities.ActivityFileTypeCoverImage, files[0].Purpose)
						assert.True(t, files[0].Public)
						return nil
					})
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "updates cover image and participant list",
			id:   "2",
			activity: entities.Activity{
				Name:            "Updated Workshop",
				CreatedBy:       "user1",
				CoverImage:      &entities.File{Name: "cubierta.jpg"},
				ParticipantList: &entities.File{Name: "lista.pdf"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).Return(nil)
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, files []*entities.File) error {
						require.Len(t, files, 2)
						assert.Equal(t, "files/activities/2/cubierta.jpg", files[0].Key)
						assert.True(t, files[0].Public)
						assert.Equal(t, "files/activities/2/participants_lista.pdf", files[1].Key)
						assert.False(t, files[1].Public)
						return nil
					})
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "storage upload error on update",
			id:   "1",
			activity: entities.Activity{
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).Return(nil)
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					Return(errors.New("storage error"))
			},
			wantErr: true,
		},
		{
			name: "save to DB error on update",
			id:   "1",
			activity: entities.Activity{
				CoverImage: &entities.File{Name: "cubierta.jpg"},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).Return(nil)
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, storage := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo, storage)
			}
			err := svc.UpdateActivity(context.Background(), tt.id, tt.activity)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ToggleFeature(t *testing.T) {
	type testCase struct {
		name      string
		id        string
		featured  bool
		prepare   func(repo *mocks.MockRepository)
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name:     "under limit succeeds",
			id:       "1",
			featured: true,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					Return(entities.Activity{ID: "1", GroupID: "g1", IsFeatured: false}, nil)
				isFeatured := true
				repo.EXPECT().CountActivities(mock.Anything, entities.ActivityFilter{GroupID: "g1", IsFeatured: &isFeatured}).
					Return(3, nil)
				repo.EXPECT().UpdateFeatureStatus(mock.Anything, "1", true).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "at limit rejected",
			id:       "1",
			featured: true,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					Return(entities.Activity{ID: "1", GroupID: "g1", IsFeatured: false}, nil)
				isFeatured := true
				repo.EXPECT().CountActivities(mock.Anything, entities.ActivityFilter{GroupID: "g1", IsFeatured: &isFeatured}).
					Return(4, nil)
			},
			wantErr:   true,
			errTarget: ErrMaxFeaturedLimitReached,
		},
		{
			name:     "re-toggling an already-featured activity at the limit succeeds (idempotent)",
			id:       "1",
			featured: true,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					Return(entities.Activity{ID: "1", GroupID: "g1", IsFeatured: true}, nil)
			},
			wantErr: false,
		},
		{
			name:     "unfeaturing does not check the limit",
			id:       "1",
			featured: false,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					Return(entities.Activity{ID: "1", GroupID: "g1", IsFeatured: true}, nil)
				repo.EXPECT().UpdateFeatureStatus(mock.Anything, "1", false).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "not found",
			id:       "99",
			featured: true,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetActivityByID(mock.Anything, "99").
					Return(entities.Activity{}, ErrActivityNotFound)
			},
			wantErr:   true,
			errTarget: ErrActivityNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo)
			}
			err := svc.ToggleFeature(context.Background(), tt.id, tt.featured)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteActivity(t *testing.T) {
	type testCase struct {
		name      string
		id        string
		prepare   func(repo *mocks.MockRepository)
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name: "success",
			id:   "1",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteActivity(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "99",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteActivity(mock.Anything, "99").
					RunAndReturn(func(_ context.Context, _ string) error {
						return ErrActivityNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrActivityNotFound,
		},
		{
			name: "db error",
			id:   "1",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteActivity(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(repo)
			}
			err := svc.DeleteActivity(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
