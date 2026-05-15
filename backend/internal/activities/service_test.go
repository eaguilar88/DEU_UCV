package activities

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/activities/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		name      string
		id        string
		prepare   func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		want      entities.Activity
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name: "success no files",
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
			want:    entities.Activity{ID: "1", Name: "Test", Files: entities.GroupedFiles{}},
			wantErr: false,
		},
		{
			name: "success with report files",
			id:   "1",
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				f := &entities.File{Key: "files/activities/1/reporte.pdf", Name: "reporte.pdf", Purpose: entities.ActivityFileTypeReport}
				repo.EXPECT().GetActivityByID(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Activity, error) {
						return entities.Activity{ID: "1", Name: "Test"}, nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeActivity).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ActivityFileTypeReport: {f}}, nil
					})
				storage.EXPECT().GetFileURL(mock.Anything, f.Key).
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "http://cdn/reporte.pdf", nil
					})
			},
			wantErr: false,
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
			got, _, err := svc.GetActivities(context.Background(), tt.filter, tt.pageScope)
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
		name      string
		activity  entities.Activity
		files     []*entities.File
		prepare   func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		wantID    int64
		wantErr   bool
	}

	tests := []testCase{
		{
			name:     "success no files",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			files:    nil,
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
			name:     "success with files",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			files: []*entities.File{
				{Name: "0_reporte.pdf", Purpose: entities.ActivityFileTypeReport},
			},
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().CreateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) (int64, error) {
						return 10, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, files []*entities.File) error {
						assert.Equal(t, "files/activities/10/0_reporte.pdf", files[0].Key)
						assert.Equal(t, entities.OwnerTypeActivity, files[0].OwnerType)
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
			files:    nil,
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
			name:     "storage upload error",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			files:    []*entities.File{{Name: "reporte.pdf"}},
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
			name:     "save files to DB error",
			activity: entities.Activity{GroupID: "1", Name: "Workshop"},
			files:    []*entities.File{{Name: "reporte.pdf"}},
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
			got, err := svc.CreateActivity(context.Background(), tt.activity, tt.files)
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
		prepare   func(repo *mocks.MockRepository)
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name:     "success",
			id:       "1",
			activity: entities.Activity{Name: "Updated Workshop"},
			prepare: func(repo *mocks.MockRepository) {
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
			prepare: func(repo *mocks.MockRepository) {
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
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().UpdateActivity(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.Activity) error {
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

