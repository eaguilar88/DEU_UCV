package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// newTestProvider returns a fresh provider with non-nil file pointers.
// Call per test run to avoid cross-test mutation of file fields.
func newTestProvider() *entities.Provider {
	return &entities.Provider{
		ID:   "provider-1",
		Type: entities.CourseProviderType,
		User: entities.User{ID: "user-1", Email: "user@test.com"},
		Files: entities.ProviderFiles{
			CI:   &entities.File{Name: "ci.pdf"},
			RIF:  &entities.File{Name: "rif.pdf"},
			ISLR: &entities.File{Name: "islr.pdf"},
		},
	}
}

func TestNewService(t *testing.T) {
	repo := mocks.NewMockRepository(t)
	storage := mocks.NewMockStorageClient(t)
	email := mocks.NewMockMailClient(t)
	logger := zap.NewNop()

	got := NewService(repo, storage, email, logger)
	assert.NotNil(t, got)
	s := got.(*service)
	assert.Equal(t, repo, s.repo)
	assert.Equal(t, storage, s.storage)
	assert.Equal(t, email, s.emailClient)
	assert.Equal(t, logger, s.logger)
}

func TestService_GetProvider(t *testing.T) {
	type testCase struct {
		name       string
		providerID string
		prepare    func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		want       entities.Provider
		wantErr    bool
		errTarget  error
	}

	tests := []testCase{
		{
			name:       "success no files",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    entities.Provider{ID: "1"},
			wantErr: false,
		},
		{
			name:       "success with CI file",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				ciFile := &entities.File{Key: "ci-key", Name: "ci.pdf"}
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {ciFile}}, nil
					})
				storageMock.EXPECT().GetFileURL(mock.Anything, "ci-key").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "http://url/ci.pdf", nil
					})
			},
			want: entities.Provider{
				ID: "1",
				Files: entities.ProviderFiles{
					CI: &entities.File{Key: "ci-key", Name: "ci.pdf", URL: "http://url/ci.pdf"},
				},
			},
			wantErr: false,
		},
		{
			name:       "provider not found",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{}, ErrProviderNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrProviderNotFound,
		},
		{
			name:       "repo error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name:       "get files error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, errors.New("fs error")
					})
			},
			wantErr: true,
		},
		{
			name:       "context cancelled",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{}, context.Canceled
					})
			},
			wantErr:   true,
			errTarget: context.Canceled,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			got, err := s.GetProvider(context.Background(), tt.providerID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repoMock.AssertExpectations(t)
			storageMock.AssertExpectations(t)
		})
	}
}

func TestService_GetProviderByCode(t *testing.T) {
	type testCase struct {
		name    string
		code    string
		prepare func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		want    entities.Provider
		wantErr bool
	}

	tests := []testCase{
		{
			name: "success",
			code: "ECP-abc123",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProviderByCode(mock.Anything, "ECP-abc123").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    entities.Provider{ID: "1"},
			wantErr: false,
		},
		{
			name: "repo error",
			code: "ECP-abc123",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProviderByCode(mock.Anything, "ECP-abc123").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{}, errors.New("not found")
					})
			},
			wantErr: true,
		},
		{
			name: "get files error",
			code: "ECP-abc123",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProviderByCode(mock.Anything, "ECP-abc123").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, errors.New("fs error")
					})
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			got, err := s.GetProviderByCode(context.Background(), tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repoMock.AssertExpectations(t)
		})
	}
}

func TestService_GetProviders(t *testing.T) {
	type testCase struct {
		name      string
		pageScope entities.PageScope
		prepare   func(repoMock *mocks.MockRepository)
		want      []entities.Provider
		wantErr   bool
	}

	tests := []testCase{
		{
			name: "success empty list",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{}, ps, nil
					})
			},
			want:    []entities.Provider{},
			wantErr: false,
		},
		{
			name: "success with one provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "1"}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    []entities.Provider{{ID: "1"}},
			wantErr: false,
		},
		{
			name: "repo error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
						return nil, entities.PageScope{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name: "get files error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "1"}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, errors.New("fs error")
					})
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			got, _, err := s.GetProviders(context.Background(), tt.pageScope)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repoMock.AssertExpectations(t)
		})
	}
}

func TestService_CreateProvider(t *testing.T) {
	type testCase struct {
		name    string
		pType   entities.ProviderType
		prepare func(repoMock *mocks.MockRepository)
		wantID  int64
		wantErr bool
	}

	tests := []testCase{
		{
			name:    "invalid provider type",
			pType:   entities.ProviderType("invalid"),
			wantID:  int64(-1),
			wantErr: true,
		},
		{
			name:  "repo create error",
			pType: entities.CourseProviderType,
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(-1), errors.New("db error")
					})
			},
			wantID:  int64(-1),
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			gotID, _, err := s.CreateProvider(context.Background(), &entities.Provider{Type: tt.pType})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantID, gotID)
			} else {
				assert.NoError(t, err)
			}
			repoMock.AssertExpectations(t)
		})
	}
}

func TestService_UpdateProvider(t *testing.T) {
	type testCase struct {
		name       string
		providerID string
		prepare    func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		wantErr    bool
	}

	tests := []testCase{
		{
			name:       "success",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().UpdateProvider(mock.Anything, "1", mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ string, _ entities.Provider) error { return nil })
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
			},
			wantErr: false,
		},
		{
			name:       "repo update error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().UpdateProvider(mock.Anything, "1", mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ string, _ entities.Provider) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name:       "invalid provider ID",
			providerID: "abc",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().UpdateProvider(mock.Anything, "abc", mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ string, _ entities.Provider) error { return nil })
			},
			wantErr: true,
		},
		{
			name:       "upload error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().UpdateProvider(mock.Anything, "1", mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ string, _ entities.Provider) error { return nil })
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("upload failed")
					})
			},
			wantErr: true,
		},
		{
			name:       "save to db error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().UpdateProvider(mock.Anything, "1", mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ string, _ entities.Provider) error { return nil })
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			err := s.UpdateProvider(context.Background(), tt.providerID, newTestProvider())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repoMock.AssertExpectations(t)
			storageMock.AssertExpectations(t)
		})
	}
}

func TestService_DeleteProvider(t *testing.T) {
	type testCase struct {
		name       string
		providerID string
		prepare    func(repoMock *mocks.MockRepository)
		wantErr    bool
	}

	tests := []testCase{
		{
			name:       "success",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().DeleteProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) error { return nil })
			},
			wantErr: false,
		},
		{
			name:       "repo error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().DeleteProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			err := s.DeleteProvider(context.Background(), tt.providerID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repoMock.AssertExpectations(t)
		})
	}
}

// TestService_uploadAndSave tests the upload-then-save path within CreateProvider.
// There is no standalone uploadAndSave method; the logic lives inline in CreateProvider
// (storage.UploadFile → repo.SaveFilesToDB → emailClient.Send).
func TestService_uploadAndSave(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient)
		wantID  int64
		wantErr bool
	}

	tests := []testCase{
		{
			name: "success",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(1), nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				mailMock.EXPECT().Send(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					RunAndReturn(func(_ context.Context, _, _, _ string) error { return nil })
			},
			wantID:  int64(1),
			wantErr: false,
		},
		{
			name: "error uploading files",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, _ *mocks.MockMailClient) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(1), nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("upload failed")
					})
			},
			wantID:  int64(-1),
			wantErr: true,
		},
		{
			name: "error saving files to db",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, _ *mocks.MockMailClient) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(1), nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("db error")
					})
			},
			wantID:  int64(-1),
			wantErr: true,
		},
		{
			name: "error sending email",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(1), nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				mailMock.EXPECT().Send(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					RunAndReturn(func(_ context.Context, _, _, _ string) error {
						return errors.New("email failed")
					})
			},
			wantID:  int64(-1),
			wantErr: true,
		},
	}

	ctx := context.Background()
	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			if tt.prepare != nil {
				tt.prepare(ctx, repoMock, storageMock, mailMock)
			}
			s := NewService(repoMock, storageMock, mailMock, loggerMock)
			gotID, gotCode, err := s.CreateProvider(ctx, newTestProvider())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(-1), gotID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
				assert.NotEmpty(t, gotCode)
			}
			repoMock.AssertExpectations(t)
			storageMock.AssertExpectations(t)
			mailMock.AssertExpectations(t)
		})
	}
}

func TestService_getFilesForProvider(t *testing.T) {
	type testCase struct {
		name       string
		providerID string
		prepare    func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		want       entities.ProviderFiles
		wantErr    bool
		errTarget  error
	}

	tests := []testCase{
		{
			name:       "success empty files",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    entities.ProviderFiles{},
			wantErr: false,
		},
		{
			name:       "success with CI file",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				ciFile := &entities.File{Key: "ci-key", Name: "ci.pdf"}
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {ciFile}}, nil
					})
				storageMock.EXPECT().GetFileURL(mock.Anything, "ci-key").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "http://url/ci.pdf", nil
					})
			},
			want: entities.ProviderFiles{
				CI: &entities.File{Key: "ci-key", Name: "ci.pdf", URL: "http://url/ci.pdf"},
			},
			wantErr: false,
		},
		{
			name:       "file not found",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, ErrFileNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrFileNotFound,
		},
		{
			name:       "repo error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name:       "get file URL error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {{Key: "ci-key"}}}, nil
					})
				storageMock.EXPECT().GetFileURL(mock.Anything, "ci-key").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "", errors.New("url error")
					})
			},
			wantErr: true,
		},
	}

	loggerMock := zap.NewNop()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			if tt.prepare != nil {
				tt.prepare(repoMock, storageMock)
			}
			s := &service{repo: repoMock, storage: storageMock, logger: loggerMock}
			got, err := s.getFilesForProvider(context.Background(), tt.providerID)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			repoMock.AssertExpectations(t)
			storageMock.AssertExpectations(t)
		})
	}
}

func Test_prepareFilesSlice(t *testing.T) {
	metadata := map[string]string{"provider_id": "42"}

	tests := []struct {
		name     string
		provider *entities.Provider
		wantLen  int
		wantKeys []string
	}{
		{
			name: "three required files",
			provider: &entities.Provider{
				User:  entities.User{ID: "user-1"},
				Files: entities.ProviderFiles{CI: &entities.File{Name: "ci.pdf"}, RIF: &entities.File{Name: "rif.pdf"}, ISLR: &entities.File{Name: "islr.pdf"}},
			},
			wantLen:  3,
			wantKeys: []string{"files/providers/42/ci.pdf", "files/providers/42/rif.pdf", "files/providers/42/islr.pdf"},
		},
		{
			name: "with resume and other",
			provider: &entities.Provider{
				User: entities.User{ID: "user-1"},
				Files: entities.ProviderFiles{
					CI:      &entities.File{Name: "ci.pdf"},
					RIF:     &entities.File{Name: "rif.pdf"},
					ISLR:    &entities.File{Name: "islr.pdf"},
					Resumes: []*entities.File{{Name: "resume.pdf"}},
					Others:  []*entities.File{{Name: "other.pdf"}},
				},
			},
			wantLen:  5,
			wantKeys: []string{"files/providers/42/ci.pdf", "files/providers/42/rif.pdf", "files/providers/42/islr.pdf", "files/providers/42/resume.pdf", "files/providers/42/other.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareFilesSlice(tt.provider, 42, metadata)
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			for i, f := range got {
				assert.Equal(t, "42", f.OwnerID)
				assert.Equal(t, entities.OwnerTypeProvider, f.OwnerType)
				assert.Equal(t, tt.wantKeys[i], f.Key)
				assert.False(t, f.Public)
				assert.Equal(t, metadata, f.MetaData)
				assert.Equal(t, "user-1", f.UploadedBy)
				assert.NotEmpty(t, f.CreatedAt)
			}
		})
	}
}

func Test_makeFileEntityFromFilePointer(t *testing.T) {
	metadata := map[string]string{"k": "v"}
	file := &entities.File{Name: "test.pdf"}

	got := makeFileEntityFromFilePointer(file, 10, "user-1", entities.OwnerTypeProvider, metadata)

	assert.Same(t, file, got)
	assert.Equal(t, "10", got.OwnerID)
	assert.Equal(t, entities.OwnerTypeProvider, got.OwnerType)
	assert.Equal(t, "files/providers/10/test.pdf", got.Key)
	assert.False(t, got.Public)
	assert.Equal(t, metadata, got.MetaData)
	assert.Equal(t, "user-1", got.UploadedBy)
	assert.NotEmpty(t, got.CreatedAt)
}
