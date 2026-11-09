package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/jwt"
	jwtMock "github.com/eaguilar88/deu/internal/jwt/mocks"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/eaguilar88/deu/internal/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	testResetURL = "https://example.test/restablecer-contrasena"
	// testPasswordHash is the bcrypt hash of "nolodire".
	testPasswordHash = "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa"
	testNewPassword  = "nuevaClave123"
)

// hashOf matches a bcrypt hash of password.
func hashOf(password string) any {
	return mock.MatchedBy(func(hash string) bool {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	})
}

func TestAuthService_Login(t *testing.T) {
	type testCase struct {
		name     string
		repo     *mocks.MockRepository
		signer   *jwtMock.MockSigner
		username string
		password string
		prepare  func(ctx context.Context, tc *testCase)
		token    string
		user     *entities.User
		wantErr  error
	}

	tests := []testCase{
		{
			name:     "success",
			username: "jon.doe@email.com",
			password: "nolodire",
			repo:     &mocks.MockRepository{},
			signer:   &jwtMock.MockSigner{},
			prepare: func(ctx context.Context, tc *testCase) {
				userRoles := []entities.UserRole{{Name: "admin", DomainType: "all", Faculty: ""}}
				tc.repo.On("GetUserByUsername", ctx, tc.username).Return(tc.user, nil)
				tc.repo.On("GetUserRoles", ctx, tc.user.ID).Return(userRoles, nil)
				tc.repo.On("GetProviderCodeByUserID", ctx, tc.user.ID).Return("", nil)
				tc.repo.On("GetGroupByUserID", ctx, tc.user.ID).Return(entities.ExtensionGroup{}, groups.ErrGroupNotFound)
				tc.repo.On("GetProviderByUserID", ctx, tc.user.ID).Return(entities.Provider{}, providers.ErrProviderNotFound)
				tc.signer.On("GenerateJWT", "1", userRoles, "", "", "", "", "").Return(tc.token, nil)
			},
			token: "token",
			user: &entities.User{
				ID:        "1",
				FirstName: "Jon",
				LastName:  "Doe",
				Password:  "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa",
				Roles: []string{
					"admin",
				},
			},
			wantErr: nil,
		},
		{
			name:     "deu_admin populates faculty from user role",
			username: "deu.admin@email.com",
			password: "nolodire",
			repo:     &mocks.MockRepository{},
			signer:   &jwtMock.MockSigner{},
			prepare: func(ctx context.Context, tc *testCase) {
				userRoles := []entities.UserRole{{Name: "deu_admin", DomainType: "all", Faculty: "DEU"}}
				tc.repo.On("GetUserByUsername", ctx, tc.username).Return(tc.user, nil)
				tc.repo.On("GetUserRoles", ctx, tc.user.ID).Return(userRoles, nil)
				tc.repo.On("GetProviderCodeByUserID", ctx, tc.user.ID).Return("", nil)
				tc.repo.On("GetGroupByUserID", ctx, tc.user.ID).Return(entities.ExtensionGroup{}, groups.ErrGroupNotFound)
				tc.repo.On("GetProviderByUserID", ctx, tc.user.ID).Return(entities.Provider{}, providers.ErrProviderNotFound)
				tc.signer.On("GenerateJWT", "2", userRoles, "", "", "", "", "").Return(tc.token, nil)
			},
			token: "token",
			user: &entities.User{
				ID:        "2",
				FirstName: "Deu",
				LastName:  "Admin",
				Password:  "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa",
				Roles: []string{
					"deu_admin",
				},
				Faculty: "DEU",
			},
			wantErr: nil,
		},
		{
			name:     "user who owns a group gets group claims",
			username: "group.owner@email.com",
			password: "nolodire",
			repo:     &mocks.MockRepository{},
			signer:   &jwtMock.MockSigner{},
			prepare: func(ctx context.Context, tc *testCase) {
				userRoles := []entities.UserRole{{Name: "extension", DomainType: "group", Faculty: ""}}
				tc.repo.On("GetUserByUsername", ctx, tc.username).Return(tc.user, nil)
				tc.repo.On("GetUserRoles", ctx, tc.user.ID).Return(userRoles, nil)
				tc.repo.On("GetProviderCodeByUserID", ctx, tc.user.ID).Return("", nil)
				tc.repo.On("GetGroupByUserID", ctx, tc.user.ID).Return(entities.ExtensionGroup{ID: "group-1", Name: "Grupo Uno"}, nil)
				tc.repo.On("GetProviderByUserID", ctx, tc.user.ID).Return(entities.Provider{}, providers.ErrProviderNotFound)
				tc.signer.On("GenerateJWT", "3", userRoles, "", "group-1", "Grupo Uno", "", "").Return(tc.token, nil)
			},
			token: "token",
			user: &entities.User{
				ID:        "3",
				FirstName: "Group",
				LastName:  "Owner",
				Password:  "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa",
				Roles: []string{
					"extension",
				},
				GroupID:   "group-1",
				GroupName: "Grupo Uno",
			},
			wantErr: nil,
		},
		{
			name:     "course provider gets provider claims",
			username: "course.provider@email.com",
			password: "nolodire",
			repo:     &mocks.MockRepository{},
			signer:   &jwtMock.MockSigner{},
			prepare: func(ctx context.Context, tc *testCase) {
				userRoles := []entities.UserRole{{Name: "coordinador", DomainType: "course", Faculty: ""}}
				tc.repo.On("GetUserByUsername", ctx, tc.username).Return(tc.user, nil)
				tc.repo.On("GetUserRoles", ctx, tc.user.ID).Return(userRoles, nil)
				tc.repo.On("GetProviderCodeByUserID", ctx, tc.user.ID).Return("ECP-abc123", nil)
				tc.repo.On("GetGroupByUserID", ctx, tc.user.ID).Return(entities.ExtensionGroup{}, groups.ErrGroupNotFound)
				tc.repo.On("GetProviderByUserID", ctx, tc.user.ID).Return(entities.Provider{ID: "provider-1", Name: "Proveedor Uno", Code: "ECP-abc123"}, nil)
				tc.signer.On("GenerateJWT", "4", userRoles, "ECP-abc123", "", "", "provider-1", "Proveedor Uno").Return(tc.token, nil)
			},
			token: "token",
			user: &entities.User{
				ID:        "4",
				FirstName: "Course",
				LastName:  "Provider",
				Password:  "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa",
				Roles: []string{
					"coordinador",
				},
				ProviderCode:       "ECP-abc123",
				CourseProviderID:   "provider-1",
				CourseProviderName: "Proveedor Uno",
			},
			wantErr: nil,
		},
		{
			name:     "group provider does not get course provider claims",
			username: "group.provider@email.com",
			password: "nolodire",
			repo:     &mocks.MockRepository{},
			signer:   &jwtMock.MockSigner{},
			prepare: func(ctx context.Context, tc *testCase) {
				userRoles := []entities.UserRole{{Name: "extension", DomainType: "group", Faculty: ""}}
				tc.repo.On("GetUserByUsername", ctx, tc.username).Return(tc.user, nil)
				tc.repo.On("GetUserRoles", ctx, tc.user.ID).Return(userRoles, nil)
				tc.repo.On("GetProviderCodeByUserID", ctx, tc.user.ID).Return("GEX-xyz789", nil)
				tc.repo.On("GetGroupByUserID", ctx, tc.user.ID).Return(entities.ExtensionGroup{}, groups.ErrGroupNotFound)
				tc.repo.On("GetProviderByUserID", ctx, tc.user.ID).Return(entities.Provider{ID: "provider-2", Name: "Proveedor Grupo", Code: "GEX-xyz789"}, nil)
				tc.signer.On("GenerateJWT", "5", userRoles, "GEX-xyz789", "", "", "", "").Return(tc.token, nil)
			},
			token: "token",
			user: &entities.User{
				ID:        "5",
				FirstName: "Group",
				LastName:  "Provider",
				Password:  "$2a$10$rFfAKJJIvbGaQCay8zC9bulGQ/kOYDOwwVBGr0WyWA1sUTLZsuXpa",
				Roles: []string{
					"extension",
				},
				ProviderCode: "GEX-xyz789",
			},
			wantErr: nil,
		},
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			s := NewService(tt.repo, tt.signer, mocks.NewMockMailClient(t), testResetURL, loggerMock)
			got, got1, err := s.Login(ctx, tt.username, tt.password)
			assert.Equal(t, tt.token, got)
			assert.Equal(t, tt.user, got1)
			assert.Equal(t, tt.wantErr, err)
			tt.repo.AssertExpectations(t)
			tt.signer.AssertExpectations(t)
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	dbErr := errors.New("db error")
	tests := []struct {
		name            string
		currentPassword string
		prepare         func(repo *mocks.MockRepository)
		wantErr         error
	}{
		{
			name:            "success",
			currentPassword: "nolodire",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(&entities.User{ID: "1", Password: testPasswordHash}, nil)
				repo.EXPECT().UpdateUserPassword(mock.Anything, "1", hashOf(testNewPassword)).Return(nil)
			},
		},
		{
			name:            "wrong current password",
			currentPassword: "incorrecta",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(&entities.User{ID: "1", Password: testPasswordHash}, nil)
			},
			wantErr: ErrInvalidCurrentPassword,
		},
		{
			name:            "error getting credentials",
			currentPassword: "nolodire",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(nil, dbErr)
			},
			wantErr: dbErr,
		},
		{
			name:            "error updating password",
			currentPassword: "nolodire",
			prepare: func(repo *mocks.MockRepository) {
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(&entities.User{ID: "1", Password: testPasswordHash}, nil)
				repo.EXPECT().UpdateUserPassword(mock.Anything, "1", mock.Anything).Return(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			tt.prepare(repo)
			s := NewService(repo, jwtMock.NewMockSigner(t), mocks.NewMockMailClient(t), testResetURL, zap.NewNop())

			err := s.ChangePassword(context.Background(), "1", tt.currentPassword, testNewPassword)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestAuthService_RequestPasswordReset(t *testing.T) {
	user := &entities.User{ID: "1", Email: "ana.perez@example.test", FirstName: "Ana", Password: testPasswordHash}
	wantData := email.PasswordResetData{Name: "Ana", ResetURL: testResetURL + "?token=reset-token"}

	t.Run("registered email sends the reset link", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		signer := jwtMock.NewMockSigner(t)
		mail := mocks.NewMockMailClient(t)
		sent := make(chan struct{})
		repo.EXPECT().GetUserByUsername(mock.Anything, user.Email).Return(user, nil)
		signer.EXPECT().GeneratePasswordResetToken("1", testPasswordHash).Return("reset-token", nil)
		mail.EXPECT().SendTemplate(mock.Anything, user.Email, email.TemplatePasswordReset, wantData).
			Run(func(context.Context, string, email.Template, any) { close(sent) }).
			Return(nil)

		err := NewService(repo, signer, mail, testResetURL, zap.NewNop()).RequestPasswordReset(context.Background(), user.Email)
		require.NoError(t, err)
		waitFor(t, sent)
	})

	t.Run("unknown email sends nothing and succeeds", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().GetUserByUsername(mock.Anything, "nadie@example.test").
			Return(nil, fmt.Errorf("%w: no rows", users.ErrUserNotFound))

		err := NewService(repo, jwtMock.NewMockSigner(t), mocks.NewMockMailClient(t), testResetURL, zap.NewNop()).
			RequestPasswordReset(context.Background(), "nadie@example.test")
		assert.NoError(t, err)
	})

	t.Run("email send failure still succeeds", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		signer := jwtMock.NewMockSigner(t)
		mail := mocks.NewMockMailClient(t)
		sent := make(chan struct{})
		repo.EXPECT().GetUserByUsername(mock.Anything, user.Email).Return(user, nil)
		signer.EXPECT().GeneratePasswordResetToken("1", testPasswordHash).Return("reset-token", nil)
		mail.EXPECT().SendTemplate(mock.Anything, user.Email, email.TemplatePasswordReset, wantData).
			Run(func(context.Context, string, email.Template, any) { close(sent) }).
			Return(errors.New("smtp down"))

		err := NewService(repo, signer, mail, testResetURL, zap.NewNop()).RequestPasswordReset(context.Background(), user.Email)
		require.NoError(t, err)
		waitFor(t, sent)
	})

	t.Run("database error is returned", func(t *testing.T) {
		dbErr := errors.New("db error")
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().GetUserByUsername(mock.Anything, user.Email).Return(nil, dbErr)

		err := NewService(repo, jwtMock.NewMockSigner(t), mocks.NewMockMailClient(t), testResetURL, zap.NewNop()).
			RequestPasswordReset(context.Background(), user.Email)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestAuthService_ResetPassword(t *testing.T) {
	user := &entities.User{ID: "1", Password: testPasswordHash}
	tests := []struct {
		name    string
		prepare func(repo *mocks.MockRepository, signer *jwtMock.MockSigner)
		wantErr error
	}{
		{
			name: "success",
			prepare: func(repo *mocks.MockRepository, signer *jwtMock.MockSigner) {
				signer.EXPECT().ValidatePasswordResetToken("reset-token").Return("1", jwt.PasswordFingerprint(testPasswordHash), nil)
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(user, nil)
				repo.EXPECT().UpdateUserPassword(mock.Anything, "1", hashOf(testNewPassword)).Return(nil)
			},
		},
		{
			name: "invalid token",
			prepare: func(_ *mocks.MockRepository, signer *jwtMock.MockSigner) {
				signer.EXPECT().ValidatePasswordResetToken("reset-token").Return("", "", jwt.ErrInvalidPasswordResetToken)
			},
			wantErr: ErrInvalidResetToken,
		},
		{
			name: "password already changed",
			prepare: func(repo *mocks.MockRepository, signer *jwtMock.MockSigner) {
				signer.EXPECT().ValidatePasswordResetToken("reset-token").Return("1", jwt.PasswordFingerprint("otro-hash"), nil)
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(user, nil)
			},
			wantErr: ErrInvalidResetToken,
		},
		{
			name: "user no longer exists",
			prepare: func(repo *mocks.MockRepository, signer *jwtMock.MockSigner) {
				signer.EXPECT().ValidatePasswordResetToken("reset-token").Return("1", jwt.PasswordFingerprint(testPasswordHash), nil)
				repo.EXPECT().GetUserCredentialsByID(mock.Anything, "1").Return(nil, users.ErrUserNotFound)
			},
			wantErr: ErrInvalidResetToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			signer := jwtMock.NewMockSigner(t)
			tt.prepare(repo, signer)
			s := NewService(repo, signer, mocks.NewMockMailClient(t), testResetURL, zap.NewNop())

			err := s.ResetPassword(context.Background(), "reset-token", testNewPassword)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// waitFor fails the test if done isn't closed soon.
func waitFor(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the email to be sent")
	}
}
