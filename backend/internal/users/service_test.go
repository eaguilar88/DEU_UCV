package users

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// newTestUser returns a fresh User entity for use in tests.
func newTestUser() *entities.User {
	return &entities.User{
		ID:          "user-1",
		CI:          "12345678",
		Email:       "test@test.com",
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: "1990-03-15",
		Gender:      "male",
		Address:     "123 Main St",
	}
}

func TestService_GetUser(t *testing.T) {
	type testCase struct {
		name      string
		userID    string
		prepare   func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		want      entities.User
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name:   "success - no profile picture",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return newTestUser(), nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{}, nil
					})
			},
			want:    *newTestUser(),
			wantErr: false,
		},
		{
			name:   "success - with profile picture",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return newTestUser(), nil
					})
				repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).
					RunAndReturn(func(_ context.Context, _ string, _ entities.OwnerType) (entities.GroupedFiles, error) {
						return entities.GroupedFiles{
							"profile_picture": {&entities.File{Key: "users/1/profile_picture.jpg"}},
						}, nil
					})
				storage.EXPECT().GetFileURL(mock.Anything, "users/1/profile_picture.jpg").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "https://cdn.example.com/users/1/profile_picture.jpg", nil
					})
			},
			want: func() entities.User {
				u := *newTestUser()
				u.ProfilePictureURL = "https://cdn.example.com/users/1/profile_picture.jpg"
				return u
			}(),
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: "unknown",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "unknown").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			email := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo, storage)

			svc := NewService(repo, email, storage, zap.NewNop())
			got, err := svc.GetUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestService_CreateUser(t *testing.T) {
	type testCase struct {
		name       string
		user       entities.User
		profilePic *entities.File
		prepare    func(repo *mocks.MockRepository, mail *mocks.MockMailClient, storage *mocks.MockStorageClient)
		wantID     int64
		wantErr    bool
		errTarget  error
	}

	tests := []testCase{
		{
			name:       "success - no profile picture",
			user:       *newTestUser(),
			profilePic: nil,
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return 1, nil
					})
				mail.EXPECT().Send(mock.Anything, "test@test.com", mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _, _, _ string) error {
						return nil
					})
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name: "success - with profile picture",
			user: *newTestUser(),
			profilePic: &entities.File{
				Name:    "profile_picture.jpg",
				Purpose: "profile_picture",
				Body:    strings.NewReader("fake image data"),
			},
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient, storage *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return 1, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return nil
					})
				repo.EXPECT().SaveFilesToDB(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return nil
					})
				mail.EXPECT().Send(mock.Anything, "test@test.com", mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _, _, _ string) error {
						return nil
					})
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name: "profile picture upload error",
			user: *newTestUser(),
			profilePic: &entities.File{
				Name:    "profile_picture.jpg",
				Purpose: "profile_picture",
				Body:    strings.NewReader("fake image data"),
			},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockMailClient, storage *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return 1, nil
					})
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ []*entities.File) error {
						return errors.New("storage error")
					})
			},
			wantErr: true,
		},
		{
			name: "profile picture db save error",
			user: *newTestUser(),
			profilePic: &entities.File{
				Name:    "profile_picture.jpg",
				Purpose: "profile_picture",
				Body:    strings.NewReader("fake image data"),
			},
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockMailClient, storage *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return 1, nil
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
			wantErr: true,
		},
		{
			name:       "user already exists",
			user:       *newTestUser(),
			profilePic: nil,
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockMailClient, _ *mocks.MockStorageClient) {
				existing := newTestUser()
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return existing, nil
					})
			},
			wantErr:   true,
			errTarget: ErrUserAlreadyExists,
		},
		{
			name:       "repo create error",
			user:       *newTestUser(),
			profilePic: nil,
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockMailClient, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return -1, errors.New("db error")
					})
			},
			wantErr: true,
		},
		{
			name:       "email send error",
			user:       *newTestUser(),
			profilePic: nil,
			prepare: func(repo *mocks.MockRepository, mail *mocks.MockMailClient, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUserByUsername(mock.Anything, "test@test.com").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
				repo.EXPECT().CreateUser(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _ entities.User) (int64, error) {
						return 1, nil
					})
				mail.EXPECT().Send(mock.Anything, "test@test.com", mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, _, _, _ string) error {
						return errors.New("mail error")
					})
			},
			wantID:  1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo, mail, storage)

			svc := NewService(repo, mail, storage, zap.NewNop())
			id, err := svc.CreateUser(context.Background(), tt.user, tt.profilePic)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
			repo.AssertExpectations(t)
			mail.AssertExpectations(t)
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	type testCase struct {
		name      string
		userID    string
		user      entities.User
		prepare   func(repo *mocks.MockRepository)
		wantErr   bool
		errTarget error
	}

	tests := []testCase{
		{
			name:   "success",
			userID: "user-1",
			user:   *newTestUser(),
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return newTestUser(), nil
					})
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", mock.Anything).
					RunAndReturn(func(_ context.Context, _ string, _ entities.User) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: "unknown",
			user:   *newTestUser(),
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUser(mock.Anything, "unknown").
					RunAndReturn(func(_ context.Context, _ string) (*entities.User, error) {
						return nil, ErrUserNotFound
					})
			},
			wantErr:   true,
			errTarget: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo)

			svc := NewService(repo, mail, storage, zap.NewNop())
			err := svc.UpdateUser(context.Background(), tt.userID, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.ErrorIs(t, err, tt.errTarget)
				}
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	type testCase struct {
		name    string
		userID  string
		prepare func(repo *mocks.MockRepository)
		wantErr bool
	}

	tests := []testCase{
		{
			name:   "success",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteUser(mock.Anything, "user-1").
					RunAndReturn(func(_ context.Context, _ string) error {
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:   "repo error",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteUser(mock.Anything, "user-1").
					RunAndReturn(func(_ context.Context, _ string) error {
						return errors.New("db error")
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo)

			svc := NewService(repo, mail, storage, zap.NewNop())
			err := svc.DeleteUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}
