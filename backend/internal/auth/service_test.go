package auth

import (
	"context"
	"testing"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	jwtMock "github.com/eaguilar88/deu/internal/jwt/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

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
				tc.signer.On("GenerateJWT", "1", userRoles, "").Return(tc.token, nil)
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
	}
	ctx := context.Background()
	loggerMock := zap.NewNop()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepare != nil {
				tt.prepare(ctx, &tt)
			}
			s := NewService(tt.repo, tt.signer, loggerMock)
			got, got1, err := s.Login(ctx, tt.username, tt.password)
			assert.Equal(t, tt.token, got)
			assert.Equal(t, tt.user, got1)
			assert.Equal(t, tt.wantErr, err)
			tt.repo.AssertExpectations(t)
			tt.signer.AssertExpectations(t)
		})
	}
}
