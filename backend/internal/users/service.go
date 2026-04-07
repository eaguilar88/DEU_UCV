package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	Send(ctx context.Context, to string, subject string, body string) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
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

func (s *service) GetUser(ctx context.Context, userID string) (entities.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return entities.User{}, err
	}
	files, err := s.repo.GetFilesByOwner(ctx, userID, entities.OwnerTypeUser)
	if err != nil {
		return entities.User{}, err
	}
	if pic := files.GetSingleFile("profile_picture"); pic != nil {
		if url, err := s.storage.GetFileURL(ctx, pic.Key); err != nil {
			s.log.Error("failed to get profile picture URL", zap.Error(err))
		} else {
			user.ProfilePictureURL = url
		}
	}
	return *user, nil
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

	if err := s.emailClient.Send(ctx, user.Email, "Welcome to DEU", "Welcome to DEU"); err != nil {
		return -1, fmt.Errorf("error sending welcome email: %w", err)
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
