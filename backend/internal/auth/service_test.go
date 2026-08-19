package auth

import (
	"context"
	"testing"

	"github.com/eaguilar88/deu/internal/auth/mocks"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups"
	jwtMock "github.com/eaguilar88/deu/internal/jwt/mocks"
	"github.com/eaguilar88/deu/internal/providers"
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
