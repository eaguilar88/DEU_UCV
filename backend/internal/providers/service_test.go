package providers

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/providers/mocks"
	"github.com/eaguilar88/deu/internal/users"
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
				repoMock.EXPECT().GetProviderContracts(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]entities.ProviderContract, error) {
						return nil, nil
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {ciFile}}, nil
					})
				storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "ci-key").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "http://url/ci.pdf", nil
					})
				repoMock.EXPECT().GetProviderContracts(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]entities.ProviderContract, error) {
						return nil, nil
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
		{
			name:       "get contracts error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
				repoMock.EXPECT().GetProviderContracts(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]entities.ProviderContract, error) {
						return nil, errors.New("db error")
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
			got, err := s.GetProvider(context.Background(), tt.providerID, entities.Viewer{UserID: "admin", Roles: []string{"deu_admin"}})
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
		filters   entities.ProviderFilters
		prepare   func(repoMock *mocks.MockRepository)
		want      []entities.Provider
		wantErr   bool
	}

	tests := []testCase{
		{
			name: "success empty list",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope, _ entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{}, ps, nil
					})
			},
			want:    []entities.Provider{},
			wantErr: false,
		},
		{
			name: "success with one provider",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope, _ entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "1"}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    []entities.Provider{{ID: "1"}},
			wantErr: false,
		},
		{
			name:    "filter by type courses",
			filters: entities.ProviderFilters{Type: entities.CourseProviderType},
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope, f entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "2", Type: entities.CourseProviderType}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "2", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    []entities.Provider{{ID: "2", Type: entities.CourseProviderType}},
			wantErr: false,
		},
		{
			name: "filter by status active",
			filters: entities.ProviderFilters{
				ProviderAdminFilters: entities.ProviderAdminFilters{
					Status: entities.ProviderStatusActive,
				},
			},
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope, f entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "3", Status: entities.ProviderStatusActive}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "3", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    []entities.Provider{{ID: "3", Status: entities.ProviderStatusActive}},
			wantErr: false,
		},
		{
			name: "repo error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.PageScope, _ entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return nil, entities.PageScope{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name: "get files error",
			prepare: func(repoMock *mocks.MockRepository) {
				repoMock.EXPECT().GetProviders(mock.Anything, mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, ps entities.PageScope, _ entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
						return []entities.Provider{{ID: "1"}}, ps, nil
					})
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
			got, _, err := s.GetProviders(context.Background(), tt.pageScope, tt.filters)
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
			gotID, err := s.CreateProvider(context.Background(), &entities.Provider{Type: tt.pType})
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
				repoMock.EXPECT().GetProviderContactInfo(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.User, error) {
						return entities.User{Email: "user@test.com", FirstName: "Test", LastName: "User"}, nil
					})
				mailMock.EXPECT().SendTemplate(mock.Anything, mock.AnythingOfType("string"), email.TemplateProviderRegistrationReceived, nil).
					RunAndReturn(func(_ context.Context, _ string, _ email.Template, _ any) error { return nil })
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
			name: "error getting contact info",
			prepare: func(ctx context.Context, repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, _ *mocks.MockMailClient) {
				repoMock.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("entities.Provider")).
					RunAndReturn(func(_ context.Context, _ entities.Provider) (int64, error) {
						return int64(1), nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().GetProviderContactInfo(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.User, error) {
						return entities.User{}, errors.New("db error")
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
				repoMock.EXPECT().GetProviderContactInfo(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.User, error) {
						return entities.User{Email: "user@test.com", FirstName: "Test", LastName: "User"}, nil
					})
				mailMock.EXPECT().SendTemplate(mock.Anything, mock.AnythingOfType("string"), email.TemplateProviderRegistrationReceived, nil).
					RunAndReturn(func(_ context.Context, _ string, _ email.Template, _ any) error {
						return errors.New("email failed")
					})
			},
			wantID:  int64(1),
			wantErr: false,
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
			gotID, err := s.CreateProvider(ctx, newTestProvider())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(-1), gotID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {ciFile}}, nil
					})
				storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "ci-key").
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
			name:       "public logo uses proxied URL",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				logo := &entities.File{Key: "logo-key", Name: "logo.png", Public: true}
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeLogo: {logo}}, nil
					})
				storageMock.EXPECT().GetFileURL(mock.Anything, "logo-key").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "https://extension.ucv.ve/files/logo-key", nil
					})
			},
			want: entities.ProviderFiles{Logo: &entities.File{
				Key: "logo-key", Name: "logo.png", Public: true, URL: "https://extension.ucv.ve/files/logo-key",
			}},
			wantErr: false,
		},
		{
			name:       "file not found",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
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
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name:       "get file URL error",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeProvider).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{entities.ProviderFileTypeCI: {{Key: "ci-key"}}}, nil
					})
				storageMock.EXPECT().GetPresignedFileURL(mock.Anything, "ci-key").
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
		// wantPublic lists the keys uploaded as public; every other file is private.
		wantPublic []string
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
		{
			name: "logo is uploaded as public",
			provider: &entities.Provider{
				User: entities.User{ID: "user-1"},
				Files: entities.ProviderFiles{
					CI:   &entities.File{Name: "ci.pdf"},
					RIF:  &entities.File{Name: "rif.pdf"},
					ISLR: &entities.File{Name: "islr.pdf"},
					Logo: &entities.File{Name: "logo.jpg", Purpose: entities.ProviderFileTypeLogo},
				},
			},
			wantLen:    4,
			wantKeys:   []string{"files/providers/42/ci.pdf", "files/providers/42/rif.pdf", "files/providers/42/islr.pdf", "files/providers/42/logo.jpg"},
			wantPublic: []string{"files/providers/42/logo.jpg"},
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
				assert.Equal(t, slices.Contains(tt.wantPublic, f.Key), f.Public, f.Key)
				assert.Equal(t, metadata, f.MetaData)
				assert.Equal(t, "user-1", f.UploadedBy)
				assert.NotEmpty(t, f.CreatedAt)
			}
		})
	}
}

func TestService_SubmitProviderContract(t *testing.T) {
	type testCase struct {
		name                                        string
		providerID                                  string
		intentionLetter, commitmentLetter, addendum *entities.File
		prepare                                     func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient)
		wantErr                                     error
		wantType                                    entities.ContractType
		wantCovered                                 []string
	}

	tests := []testCase{
		{
			name:             "initial contract, covers uncovered courses",
			providerID:       "1",
			intentionLetter:  &entities.File{Name: "carta_intencion.pdf"},
			commitmentLetter: &entities.File{Name: "carta_compromiso.pdf"},
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1", User: entities.User{ID: "user-1"}}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return false, nil })
				repoMock.EXPECT().GetUncoveredCourseIDs(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]string, error) { return []string{"10", "11"}, nil })
				repoMock.EXPECT().CreateProviderContract(mock.Anything, mock.AnythingOfType("entities.ProviderContract"), []string{"10", "11"}).
					RunAndReturn(func(_ context.Context, c entities.ProviderContract, ids []string) (entities.ProviderContract, error) {
						c.ID = "100"
						c.CoveredCourses = ids
						return c, nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
			},
			wantType:    entities.ContractTypeInitial,
			wantCovered: []string{"10", "11"},
		},
		{
			name:             "initial contract, zero uncovered courses is allowed",
			providerID:       "1",
			intentionLetter:  &entities.File{Name: "carta_intencion.pdf"},
			commitmentLetter: &entities.File{Name: "carta_compromiso.pdf"},
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1", User: entities.User{ID: "user-1"}}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return false, nil })
				repoMock.EXPECT().GetUncoveredCourseIDs(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]string, error) { return nil, nil })
				repoMock.EXPECT().CreateProviderContract(mock.Anything, mock.AnythingOfType("entities.ProviderContract"), []string(nil)).
					RunAndReturn(func(_ context.Context, c entities.ProviderContract, ids []string) (entities.ProviderContract, error) {
						c.ID = "100"
						return c, nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
			},
			wantType: entities.ContractTypeInitial,
		},
		{
			name:            "initial contract missing commitment letter",
			providerID:      "1",
			intentionLetter: &entities.File{Name: "carta_intencion.pdf"},
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return false, nil })
			},
			wantErr: ErrMissingInitialContract,
		},
		{
			name:       "addendum, covers uncovered courses",
			providerID: "1",
			addendum:   &entities.File{Name: "adenda.pdf"},
			prepare: func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1", User: entities.User{ID: "user-1"}}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return true, nil })
				repoMock.EXPECT().GetUncoveredCourseIDs(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]string, error) { return []string{"12"}, nil })
				repoMock.EXPECT().CreateProviderContract(mock.Anything, mock.AnythingOfType("entities.ProviderContract"), []string{"12"}).
					RunAndReturn(func(_ context.Context, c entities.ProviderContract, ids []string) (entities.ProviderContract, error) {
						c.ID = "101"
						c.CoveredCourses = ids
						return c, nil
					})
				storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
				repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error { return nil })
			},
			wantType:    entities.ContractTypeAddendum,
			wantCovered: []string{"12"},
		},
		{
			name:       "addendum with nothing to cover is rejected",
			providerID: "1",
			addendum:   &entities.File{Name: "adenda.pdf"},
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return true, nil })
				repoMock.EXPECT().GetUncoveredCourseIDs(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) ([]string, error) { return nil, nil })
			},
			wantErr: ErrNoCoursesToCoverage,
		},
		{
			name:       "addendum missing file",
			providerID: "1",
			prepare: func(repoMock *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repoMock.EXPECT().GetProvider(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (entities.Provider, error) {
						return entities.Provider{ID: "1"}, nil
					})
				repoMock.EXPECT().HasInitialContract(mock.Anything, "1").
					RunAndReturn(func(_ context.Context, _ string) (bool, error) { return true, nil })
			},
			wantErr: ErrMissingAddendum,
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
			wantErr: ErrProviderNotFound,
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
			got, err := s.SubmitProviderContract(context.Background(), tt.providerID, tt.intentionLetter, tt.commitmentLetter, tt.addendum)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantType, got.Type)
				assert.Equal(t, tt.wantCovered, got.CoveredCourses)
			}
			repoMock.AssertExpectations(t)
			storageMock.AssertExpectations(t)
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

func TestService_GetProvider_Access(t *testing.T) {
	provider := entities.Provider{ID: "5", User: entities.User{ID: "10"}, Faculty: entities.FacultyIngenieria}

	tests := []struct {
		name    string
		viewer  entities.Viewer
		allowed bool
	}{
		{name: "anonymous", viewer: entities.Viewer{}, allowed: false},
		{name: "provider's own user", viewer: entities.Viewer{UserID: "10"}, allowed: true},
		{name: "root", viewer: entities.Viewer{UserID: "1", Roles: []string{"root"}}, allowed: true},
		{
			name:    "faculty_admin of the provider's faculty",
			viewer:  entities.Viewer{UserID: "3", Roles: []string{"faculty_admin"}, Faculty: entities.FacultyIngenieria},
			allowed: true,
		},
		{
			name:    "faculty_admin of another faculty",
			viewer:  entities.Viewer{UserID: "4", Roles: []string{"faculty_admin"}, Faculty: entities.FacultyCiencias},
			allowed: false,
		},
		{name: "group_admin", viewer: entities.Viewer{UserID: "6", Roles: []string{"group_admin"}}, allowed: false},
		{name: "other provider", viewer: entities.Viewer{UserID: "11", Roles: []string{"course_admin"}}, allowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			repoMock.EXPECT().GetProvider(mock.Anything, "5").Return(provider, nil)
			if tt.allowed {
				repoMock.EXPECT().GetFilesByOwner(mock.Anything, "5", entities.OwnerTypeProvider).
					Return(entities.GroupedFiles{}, nil)
				repoMock.EXPECT().GetProviderContracts(mock.Anything, "5").
					Return([]entities.ProviderContract{}, nil)
			}
			// When denied, no file lookups or URL generation happen: the mocks fail on unexpected calls.

			_, err := NewService(repoMock, storageMock, mocks.NewMockMailClient(t), zap.NewNop()).
				GetProvider(context.Background(), "5", tt.viewer)

			if tt.allowed {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrProviderForbidden)
			}
		})
	}
}

func TestService_ApproveProvider(t *testing.T) {
	visitante := entities.RoleNameFromID(entities.RoleVisitante)
	courseAdmin := entities.RoleNameFromID(entities.RoleCourseAdmin)
	pending := entities.Provider{
		ID: "7", Name: "Academia Ficticia", Status: entities.ProviderStatusUnderReview,
		User: entities.User{ID: "15", Email: "proveedor@example.test"},
	}
	approvedEmail := email.ProviderApprovedData{ProviderName: "Academia Ficticia"}

	t.Run("promotes the provider's user and emails them in the transaction", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		mailMock := mocks.NewMockMailClient(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(pending, nil)
		repoMock.EXPECT().ApproveProvider(mock.Anything, "7", "15", visitante, courseAdmin, mock.Anything).
			RunAndReturn(func(_ context.Context, _, _, _, _ string, notify func() error) error { return notify() })
		mailMock.EXPECT().SendTemplate(mock.Anything, "proveedor@example.test", email.TemplateProviderApproved, approvedEmail).Return(nil)

		s := &service{repo: repoMock, emailClient: mailMock, logger: zap.NewNop()}
		assert.NoError(t, s.ApproveProvider(context.Background(), "7"))
	})

	t.Run("a failed email fails the approval", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		mailMock := mocks.NewMockMailClient(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(pending, nil)
		repoMock.EXPECT().ApproveProvider(mock.Anything, "7", "15", visitante, courseAdmin, mock.Anything).
			RunAndReturn(func(_ context.Context, _, _, _, _ string, notify func() error) error { return notify() })
		mailMock.EXPECT().SendTemplate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp down"))

		s := &service{repo: repoMock, emailClient: mailMock, logger: zap.NewNop()}
		assert.EqualError(t, s.ApproveProvider(context.Background(), "7"), "error sending provider notification email: smtp down")
	})

	t.Run("a role that was not updated is reported", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(pending, nil)
		// The repository rolls back the approval when the role is not updated.
		repoMock.EXPECT().ApproveProvider(mock.Anything, "7", "15", visitante, courseAdmin, mock.Anything).Return(users.ErrUserRoleNotFound)

		s := &service{repo: repoMock, emailClient: mocks.NewMockMailClient(t), logger: zap.NewNop()}
		assert.ErrorIs(t, s.ApproveProvider(context.Background(), "7"), users.ErrUserRoleNotFound)
	})

	t.Run("unknown provider is not approved", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(entities.Provider{}, ErrProviderNotFound)

		s := &service{repo: repoMock, emailClient: mocks.NewMockMailClient(t), logger: zap.NewNop()}
		assert.ErrorIs(t, s.ApproveProvider(context.Background(), "7"), ErrProviderNotFound)
	})

	t.Run("an already reviewed provider is not approved again", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		active := pending
		active.Status = entities.ProviderStatusActive
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(active, nil)

		s := &service{repo: repoMock, emailClient: mocks.NewMockMailClient(t), logger: zap.NewNop()}
		assert.ErrorIs(t, s.ApproveProvider(context.Background(), "7"), ErrProviderAlreadyProcessed)
	})
}

func TestService_RejectProvider(t *testing.T) {
	pending := entities.Provider{
		ID: "7", Status: entities.ProviderStatusUnderReview,
		// Without a provider name, the email uses the user's full name.
		User: entities.User{ID: "15", Email: "proveedor@example.test", FirstName: "Ana", LastName: "Pérez"},
	}

	t.Run("rejects and emails the reason in the transaction", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		mailMock := mocks.NewMockMailClient(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(pending, nil)
		repoMock.EXPECT().RejectProvider(mock.Anything, "7", mock.Anything).
			RunAndReturn(func(_ context.Context, _ string, notify func() error) error { return notify() })
		mailMock.EXPECT().SendTemplate(mock.Anything, "proveedor@example.test", email.TemplateProviderRejected,
			email.ProviderRejectedData{ProviderName: "Ana Pérez", Reason: "Falta el RIF vigente"}).Return(nil)

		s := &service{repo: repoMock, emailClient: mailMock, logger: zap.NewNop()}
		assert.NoError(t, s.RejectProvider(context.Background(), "7", "Falta el RIF vigente"))
	})

	t.Run("repository errors are returned", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(pending, nil)
		repoMock.EXPECT().RejectProvider(mock.Anything, "7", mock.Anything).Return(errors.New("db error"))

		s := &service{repo: repoMock, emailClient: mocks.NewMockMailClient(t), logger: zap.NewNop()}
		assert.EqualError(t, s.RejectProvider(context.Background(), "7", ""), "db error")
	})

	t.Run("an already reviewed provider is not rejected again", func(t *testing.T) {
		repoMock := mocks.NewMockRepository(t)
		rejected := pending
		rejected.Status = entities.ProviderStatusRejected
		repoMock.EXPECT().GetProvider(mock.Anything, "7").Return(rejected, nil)

		s := &service{repo: repoMock, emailClient: mocks.NewMockMailClient(t), logger: zap.NewNop()}
		assert.ErrorIs(t, s.RejectProvider(context.Background(), "7", ""), ErrProviderAlreadyProcessed)
	})
}

func TestService_CreateProvider_NotifiesCoordinators(t *testing.T) {
	newProvider := func() *entities.Provider {
		provider := newTestProvider()
		provider.Name = "ACME"
		provider.Faculty = entities.FacultyCiencias
		return provider
	}
	prepareCreated := func(repoMock *mocks.MockRepository, storageMock *mocks.MockStorageClient, mailMock *mocks.MockMailClient) {
		repoMock.EXPECT().CreateProvider(mock.Anything, mock.Anything).Return(int64(7), nil)
		storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
		repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
		repoMock.EXPECT().GetProviderContactInfo(mock.Anything, "7").Return(entities.User{Email: "user@test.com"}, nil)
		mailMock.EXPECT().SendTemplate(mock.Anything, "user@test.com", email.TemplateProviderRegistrationReceived, nil).Return(nil)
	}

	tests := []struct {
		name    string
		prepare func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient)
	}{
		{
			name: "sends one email per coordinator of the provider's faculty",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, entities.FacultyCiencias).
					Return([]string{"coord1@test.com", "coord2@test.com"}, nil)
				data := email.ProviderRegistrationSubmittedData{
					ProviderName: "ACME",
					ProviderType: string(entities.CourseProviderType),
					Faculty:      string(entities.FacultyCiencias),
				}
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord1@test.com", email.TemplateProviderRegistrationSubmitted, data).Return(nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord2@test.com", email.TemplateProviderRegistrationSubmitted, data).Return(nil)
			},
		},
		{
			name: "a failing coordinator email does not fail the registration",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return([]string{"coord@test.com"}, nil)
				mailMock.EXPECT().SendTemplate(mock.Anything, "coord@test.com", email.TemplateProviderRegistrationSubmitted, mock.Anything).
					Return(errors.New("smtp down"))
			},
		},
		{
			name: "a coordinator lookup error does not fail the registration",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
		},
		{
			name: "no coordinators sends nothing",
			prepare: func(repoMock *mocks.MockRepository, mailMock *mocks.MockMailClient) {
				repoMock.EXPECT().GetFacultyCoordinatorEmails(mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewMockRepository(t)
			storageMock := mocks.NewMockStorageClient(t)
			mailMock := mocks.NewMockMailClient(t)
			prepareCreated(repoMock, storageMock, mailMock)
			tt.prepare(repoMock, mailMock)

			s := NewService(repoMock, storageMock, mailMock, zap.NewNop())
			gotID, err := s.CreateProvider(context.Background(), newProvider())

			assert.NoError(t, err)
			assert.Equal(t, int64(7), gotID)
		})
	}
}

func TestService_CreateProvider_WithoutFacultySkipsCoordinators(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	storageMock := mocks.NewMockStorageClient(t)
	mailMock := mocks.NewMockMailClient(t)
	repoMock.EXPECT().CreateProvider(mock.Anything, mock.Anything).Return(int64(7), nil)
	storageMock.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
	repoMock.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).Return(nil)
	repoMock.EXPECT().GetProviderContactInfo(mock.Anything, "7").Return(entities.User{Email: "user@test.com"}, nil)
	mailMock.EXPECT().SendTemplate(mock.Anything, "user@test.com", email.TemplateProviderRegistrationReceived, nil).Return(nil)

	s := NewService(repoMock, storageMock, mailMock, zap.NewNop())
	_, err := s.CreateProvider(context.Background(), newTestProvider())

	assert.NoError(t, err)
}
