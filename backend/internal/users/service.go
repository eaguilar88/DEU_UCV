package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	UpdateUser(ctx context.Context, userID string, user entities.User) error
	DeleteUser(ctx context.Context, userID string) error
	AddRoleToUser(ctx context.Context, tx *sql.Tx, userID string, role int) error
	GetUserRoles(ctx context.Context, userID string) ([]entities.UserRole, error)
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, files []*entities.File) error
}

type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetPresignedFileURL(ctx context.Context, objectKey string) (string, error)
}

type service struct {
	repo        Repository
	emailClient MailClient
	storage     StorageClient
	log         *zap.Logger
}

func NewService(repository Repository, emailClient MailClient, storage StorageClient, logger *zap.Logger) Service {
	return &service{
		repo:        repository,
		emailClient: emailClient,
		storage:     storage,
		log:         logger,
	}
}

// GetUser returns the full user to the user themself and to admins (root, deu_admin,
// faculty_admin) and only the public profile to everyone else.
func (s *service) GetUser(ctx context.Context, userID string, viewer entities.Viewer) (entities.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return entities.User{}, err
	}
	files, err := s.repo.GetFilesByOwner(ctx, userID, entities.OwnerTypeUser)
	if err != nil {
		return entities.User{}, err
	}
	if pic := files.GetSingleFile("profile_picture"); pic != nil {
		if url, err := s.storage.GetPresignedFileURL(ctx, pic.Key); err != nil {
			s.log.Error("failed to get profile picture URL", zap.Error(err))
		} else {
			user.ProfilePictureURL = url
		}
	}
	if !canViewFullUser(userID, viewer) {
		return publicProfile(*user), nil
	}
	return *user, nil
}

func canViewFullUser(userID string, viewer entities.Viewer) bool {
	if viewer.IsAnonymous() {
		return false
	}
	return viewer.UserID == userID || viewer.IsGlobalAdmin() || viewer.IsFacultyAdmin()
}

// publicProfile keeps only what may be shown to anyone: no cédula, email, date of birth,
// address or other personal data.
func publicProfile(user entities.User) entities.User {
	return entities.User{
		ID:                user.ID,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		ProfilePictureURL: user.ProfilePictureURL,
	}
}

func (s *service) GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error) {
	users, page, err := s.repo.GetUsers(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return users, page, nil
}

func (s *service) CreateUser(ctx context.Context, user entities.User, profilePic *entities.File) (int64, error) {
	existingUser, err := s.repo.GetUserByUsername(ctx, user.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return -1, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil && existingUser.ID != "" {
		return -1, ErrUserAlreadyExists
	}

	user.Roles = []string{"visitante"}
	id, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return -1, fmt.Errorf("error creating new user: %w", err)
	}

	if profilePic != nil {
		profilePic.OwnerID = fmt.Sprintf("%d", id)
		profilePic.OwnerType = entities.OwnerTypeUser
		profilePic.Key = fmt.Sprintf("users/%d/%s", id, profilePic.Name)
		profilePic.Public = false
		profilePic.UploadedBy = fmt.Sprintf("%d", id)
		if err := s.storage.UploadFile(ctx, []*entities.File{profilePic}); err != nil {
			return -1, fmt.Errorf("error uploading profile picture: %w", err)
		}
		if err := s.repo.SaveFilesToDB(ctx, []*entities.File{profilePic}); err != nil {
			return -1, fmt.Errorf("error saving profile picture to DB: %w", err)
		}
	}

	if err := s.emailClient.SendTemplate(ctx, user.Email, email.TemplateUserWelcome, nil); err != nil {
		s.log.Warn("failed to send welcome email", zap.Error(err))
	}
	return id, nil
}

func (s *service) UpdateUser(ctx context.Context, userID string, user entities.User) error {
	userFromDB, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.CI = userFromDB.CI
	user.Email = userFromDB.Email
	user.Password = userFromDB.Password

	return s.repo.UpdateUser(ctx, userID, user)
}

func (s *service) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.DeleteUser(ctx, userID)
}
