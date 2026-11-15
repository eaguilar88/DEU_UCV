package users

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/eaguilar88/deu/internal/email"
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
				repo.EXPECT().GetUserRoles(mock.Anything, "user-1").Return([]entities.UserRole{{Name: "visitante"}}, nil)
			},
			want: func() entities.User {
				u := *newTestUser()
				u.Roles = []string{"visitante"}
				return u
			}(),
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
				storage.EXPECT().GetPresignedFileURL(mock.Anything, "users/1/profile_picture.jpg").
					RunAndReturn(func(_ context.Context, _ string) (string, error) {
						return "https://cdn.example.com/users/1/profile_picture.jpg", nil
					})
				repo.EXPECT().GetUserRoles(mock.Anything, "user-1").Return([]entities.UserRole{{Name: "visitante"}}, nil)
			},
			want: func() entities.User {
				u := *newTestUser()
				u.ProfilePictureURL = "https://cdn.example.com/users/1/profile_picture.jpg"
				u.Roles = []string{"visitante"}
				return u
			}(),
			wantErr: false,
		},
		{
			name:   "success - faculty from the coordinator role",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").Return(newTestUser(), nil)
				repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).Return(entities.GroupedFiles{}, nil)
				repo.EXPECT().GetUserRoles(mock.Anything, "user-1").Return([]entities.UserRole{
					{Name: "visitante"},
					{Name: "faculty_admin", Faculty: string(entities.FacultyCiencias)},
				}, nil)
			},
			want: func() entities.User {
				u := *newTestUser()
				u.Roles = []string{"visitante", "faculty_admin"}
				u.Faculty = string(entities.FacultyCiencias)
				return u
			}(),
			wantErr: false,
		},
		{
			name:   "error getting roles",
			userID: "user-1",
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").Return(newTestUser(), nil)
				repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).Return(entities.GroupedFiles{}, nil)
				repo.EXPECT().GetUserRoles(mock.Anything, "user-1").Return(nil, errors.New("db error"))
			},
			wantErr: true,
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
			got, err := svc.GetUser(context.Background(), tt.userID, entities.Viewer{UserID: tt.userID})

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
				mail.EXPECT().SendTemplate(mock.Anything, "test@test.com", email.TemplateUserWelcome, nil).
					RunAndReturn(func(_ context.Context, _ string, _ email.Template, _ any) error {
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
				mail.EXPECT().SendTemplate(mock.Anything, "test@test.com", email.TemplateUserWelcome, nil).
					RunAndReturn(func(_ context.Context, _ string, _ email.Template, _ any) error {
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
				mail.EXPECT().SendTemplate(mock.Anything, "test@test.com", email.TemplateUserWelcome, nil).
					RunAndReturn(func(_ context.Context, _ string, _ email.Template, _ any) error {
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

var (
	viewerSelf         = entities.Viewer{UserID: "user-1", Roles: []string{"visitante"}}
	viewerOtherUser    = entities.Viewer{UserID: "user-2", Roles: []string{"visitante"}}
	viewerDeuAdmin     = entities.Viewer{UserID: "99", Roles: []string{"deu_admin"}}
	viewerRoot         = entities.Viewer{UserID: "1", Roles: []string{"root"}}
	viewerFacultyAdmin = entities.Viewer{UserID: "50", Roles: []string{"faculty_admin"}, Faculty: entities.FacultyDEU}
)

func TestService_UpdateUser(t *testing.T) {
	str := func(s string) *string { return &s }
	pictures := func(keys ...string) entities.GroupedFiles {
		files := entities.GroupedFiles{}
		for _, k := range keys {
			files["profile_picture"] = append(files["profile_picture"], &entities.File{Key: k, Purpose: "profile_picture"})
		}
		return files
	}
	// expectUserReads covers the existence check and the final read-back of the updated user.
	expectUserReads := func(repo *mocks.MockRepository, storage *mocks.MockStorageClient, files entities.GroupedFiles) {
		repo.EXPECT().GetUser(mock.Anything, "user-1").Return(newTestUser(), nil)
		repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).Return(files, nil)
		repo.EXPECT().GetUserRoles(mock.Anything, "user-1").Return([]entities.UserRole{{Name: "visitante"}}, nil)
		storage.EXPECT().GetPresignedFileURL(mock.Anything, mock.Anything).Return("https://example.com/pic", nil).Maybe()
	}
	newPicture := func() *entities.File {
		return &entities.File{Name: "profile_picture.jpg", Purpose: "profile_picture", Body: strings.NewReader("img")}
	}
	isNewPicture := func(uploadedBy string) func(f *entities.File) bool {
		return func(f *entities.File) bool {
			return f != nil && f.Key == "users/user-1/profile_picture.jpg" && f.OwnerID == "user-1" &&
				f.OwnerType == entities.OwnerTypeUser && !f.Public && f.UploadedBy == uploadedBy
		}
	}
	changes := entities.UserChanges{FirstName: str("María"), Email: str("maria@example.com")}

	tests := []struct {
		name       string
		viewer     entities.Viewer
		changes    entities.UserChanges
		profilePic *entities.File
		prepare    func(repo *mocks.MockRepository, storage *mocks.MockStorageClient)
		errTarget  error
	}{
		{
			name:    "user updates their own profile",
			viewer:  viewerSelf,
			changes: changes,
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, entities.GroupedFiles{})
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", changes, (*entities.File)(nil)).Return(nil)
			},
		},
		{
			name:    "deu_admin updates another user",
			viewer:  viewerDeuAdmin,
			changes: changes,
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, entities.GroupedFiles{})
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", changes, (*entities.File)(nil)).Return(nil)
			},
		},
		{
			name:    "root updates another user",
			viewer:  viewerRoot,
			changes: changes,
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, entities.GroupedFiles{})
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", changes, (*entities.File)(nil)).Return(nil)
			},
		},
		{
			name:      "another user is forbidden",
			viewer:    viewerOtherUser,
			changes:   changes,
			prepare:   func(*mocks.MockRepository, *mocks.MockStorageClient) {},
			errTarget: ErrUserForbidden,
		},
		{
			name:      "faculty_admin is forbidden",
			viewer:    viewerFacultyAdmin,
			changes:   changes,
			prepare:   func(*mocks.MockRepository, *mocks.MockStorageClient) {},
			errTarget: ErrUserForbidden,
		},
		{
			name:      "anonymous is forbidden",
			viewer:    entities.Viewer{},
			changes:   changes,
			prepare:   func(*mocks.MockRepository, *mocks.MockStorageClient) {},
			errTarget: ErrUserForbidden,
		},
		{
			name:      "nothing to update",
			viewer:    viewerSelf,
			changes:   entities.UserChanges{},
			prepare:   func(*mocks.MockRepository, *mocks.MockStorageClient) {},
			errTarget: ErrInvalidInput,
		},
		{
			name:    "user not found",
			viewer:  viewerDeuAdmin,
			changes: changes,
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").Return(nil, ErrUserNotFound)
			},
			errTarget: ErrUserNotFound,
		},
		{
			name:    "email or cedula already in use",
			viewer:  viewerSelf,
			changes: changes,
			prepare: func(repo *mocks.MockRepository, _ *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").Return(newTestUser(), nil)
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", changes, (*entities.File)(nil)).
					Return(ErrUserAlreadyExists)
			},
			errTarget: ErrUserAlreadyExists,
		},
		{
			name:       "picture only, replacing one with another extension",
			viewer:     viewerDeuAdmin,
			profilePic: newPicture(),
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, pictures("users/user-1/profile_picture.png"))
				storage.EXPECT().UploadFile(mock.Anything, mock.MatchedBy(func(files []*entities.File) bool {
					return len(files) == 1 && isNewPicture("99")(files[0])
				})).Return(nil)
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", entities.UserChanges{}, mock.MatchedBy(isNewPicture("99"))).
					Return(nil)
				storage.EXPECT().DeleteFile(mock.Anything, "users/user-1/profile_picture.png").Return(nil)
			},
		},
		{
			name:       "picture with the same key is overwritten, not deleted",
			viewer:     viewerSelf,
			changes:    changes,
			profilePic: newPicture(),
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, pictures("users/user-1/profile_picture.jpg"))
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", changes, mock.MatchedBy(isNewPicture("user-1"))).
					Return(nil)
			},
		},
		{
			name:       "failing to delete the old picture does not fail the update",
			viewer:     viewerSelf,
			profilePic: newPicture(),
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				expectUserReads(repo, storage, pictures("users/user-1/profile_picture.png"))
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().UpdateUser(mock.Anything, "user-1", entities.UserChanges{}, mock.Anything).Return(nil)
				storage.EXPECT().DeleteFile(mock.Anything, "users/user-1/profile_picture.png").
					Return(errors.New("b2 error"))
			},
		},
		{
			name:       "upload failure does not touch the database",
			viewer:     viewerSelf,
			changes:    changes,
			profilePic: newPicture(),
			prepare: func(repo *mocks.MockRepository, storage *mocks.MockStorageClient) {
				repo.EXPECT().GetUser(mock.Anything, "user-1").Return(newTestUser(), nil)
				repo.EXPECT().GetFilesByOwner(mock.Anything, "user-1", entities.OwnerTypeUser).Return(entities.GroupedFiles{}, nil)
				storage.EXPECT().UploadFile(mock.Anything, mock.Anything).Return(errors.New("b2 error"))
			},
			errTarget: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo, storage)

			svc := NewService(repo, mail, storage, zap.NewNop())
			user, err := svc.UpdateUser(context.Background(), "user-1", tt.changes, tt.profilePic, tt.viewer)

			if tt.errTarget != nil {
				assert.Error(t, err)
				if tt.errTarget != errAny {
					assert.ErrorIs(t, err, tt.errTarget)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "user-1", user.ID)
			assert.Equal(t, []string{"visitante"}, user.Roles, "the read-back is the full profile")
		})
	}
}

// errAny marks a test case that expects some error, without checking which.
var errAny = errors.New("any error")

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		viewer    entities.Viewer
		prepare   func(repo *mocks.MockRepository)
		errTarget error
	}{
		{
			name:   "deu_admin deletes a user",
			viewer: viewerDeuAdmin,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteUser(mock.Anything, "user-1").Return(nil)
			},
		},
		{
			name:   "root deletes a user",
			viewer: viewerRoot,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteUser(mock.Anything, "user-1").Return(nil)
			},
		},
		{
			name:   "repo error",
			viewer: viewerDeuAdmin,
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().DeleteUser(mock.Anything, "user-1").Return(ErrUserNotFound)
			},
			errTarget: ErrUserNotFound,
		},
		{
			name:      "the user themself is forbidden",
			viewer:    viewerSelf,
			prepare:   func(*mocks.MockRepository) {},
			errTarget: ErrUserForbidden,
		},
		{
			name:      "another user is forbidden",
			viewer:    viewerOtherUser,
			prepare:   func(*mocks.MockRepository) {},
			errTarget: ErrUserForbidden,
		},
		{
			name:      "faculty_admin is forbidden",
			viewer:    viewerFacultyAdmin,
			prepare:   func(*mocks.MockRepository) {},
			errTarget: ErrUserForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			mail := mocks.NewMockMailClient(t)
			storage := mocks.NewMockStorageClient(t)
			tt.prepare(repo)

			svc := NewService(repo, mail, storage, zap.NewNop())
			err := svc.DeleteUser(context.Background(), "user-1", tt.viewer)

			if tt.errTarget != nil {
				assert.ErrorIs(t, err, tt.errTarget)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestService_GetUser_Visibility(t *testing.T) {
	fullUser := &entities.User{
		ID: "1", CI: "V123", Email: "ana@ucv.ve", FirstName: "Ana", LastName: "Pérez",
		DateOfBirth: "2000-01-01", Address: "Caracas",
	}

	tests := []struct {
		name     string
		viewer   entities.Viewer
		wantFull bool
	}{
		{name: "anonymous", viewer: entities.Viewer{}, wantFull: false},
		{name: "other user", viewer: entities.Viewer{UserID: "2", Roles: []string{"group_admin"}}, wantFull: false},
		{name: "self", viewer: entities.Viewer{UserID: "1"}, wantFull: true},
		{name: "deu_admin", viewer: entities.Viewer{UserID: "9", Roles: []string{"deu_admin"}}, wantFull: true},
		{name: "faculty_admin", viewer: entities.Viewer{UserID: "8", Roles: []string{"faculty_admin"}}, wantFull: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			storage := mocks.NewMockStorageClient(t)
			userCopy := *fullUser
			repo.EXPECT().GetUser(mock.Anything, "1").Return(&userCopy, nil)
			repo.EXPECT().GetFilesByOwner(mock.Anything, "1", entities.OwnerTypeUser).Return(entities.GroupedFiles{}, nil)
			if tt.wantFull {
				// Only the full profile carries the user's roles.
				repo.EXPECT().GetUserRoles(mock.Anything, "1").Return([]entities.UserRole{{Name: "course_admin"}}, nil)
			}

			got, err := NewService(repo, mocks.NewMockMailClient(t), storage, zap.NewNop()).
				GetUser(context.Background(), "1", tt.viewer)

			assert.NoError(t, err)
			if tt.wantFull {
				want := *fullUser
				want.Roles = []string{"course_admin"}
				assert.Equal(t, want, got)
				return
			}
			assert.Equal(t, entities.User{ID: "1", FirstName: "Ana", LastName: "Pérez"}, got)
		})
	}
}
