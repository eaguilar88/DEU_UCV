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
	UpdateUser(ctx context.Context, userID string, changes entities.UserChanges, profilePic *entities.File) error
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
	DeleteFile(ctx context.Context, objectKey string) error
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

	// The full profile carries the user's roles, which the frontends use to decide what to show.
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user roles", zap.Error(err), zap.String("user_id", userID))
		return entities.User{}, err
	}
	user.Roles = make([]string, 0, len(roles))
	for _, role := range roles {
		user.Roles = append(user.Roles, role.Name)
		if user.Faculty == "" && role.Faculty != "" {
			user.Faculty = role.Faculty
		}
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

// canEditUser reports whether the viewer may modify the user: the user themself, root or
// deu_admin.
func canEditUser(userID string, viewer entities.Viewer) bool {
	if viewer.IsAnonymous() {
		return false
	}
	return viewer.UserID == userID || viewer.IsGlobalAdmin()
}

// UpdateUser applies a partial update to the user's profile and, when profilePic is given,
// replaces their profile picture. It returns the updated user as the viewer sees it.
func (s *service) UpdateUser(
	ctx context.Context,
	userID string,
	changes entities.UserChanges,
	profilePic *entities.File,
	viewer entities.Viewer,
) (entities.User, error) {
	if !canEditUser(userID, viewer) {
		return entities.User{}, ErrUserForbidden
	}
	if changes.IsEmpty() && profilePic == nil {
		return entities.User{}, fmt.Errorf("%w: no fields to update", ErrInvalidInput)
	}

	if _, err := s.repo.GetUser(ctx, userID); err != nil {
		return entities.User{}, err
	}

	var oldPictures []*entities.File
	if profilePic != nil {
		files, err := s.repo.GetFilesByOwner(ctx, userID, entities.OwnerTypeUser)
		if err != nil {
			return entities.User{}, err
		}
		oldPictures = files.GetMultipleFiles("profile_picture")

		profilePic.OwnerID = userID
		profilePic.OwnerType = entities.OwnerTypeUser
		profilePic.Key = fmt.Sprintf("users/%s/%s", userID, profilePic.Name)
		profilePic.Public = false
		profilePic.UploadedBy = viewer.UserID
		if err := s.storage.UploadFile(ctx, []*entities.File{profilePic}); err != nil {
			s.log.Error("failed to upload profile picture", zap.Error(err), zap.String("user_id", userID))
			return entities.User{}, fmt.Errorf("error uploading profile picture: %w", err)
		}
	}

	if err := s.repo.UpdateUser(ctx, userID, changes, profilePic); err != nil {
		s.log.Error("failed to update user", zap.Error(err), zap.String("user_id", userID))
		return entities.User{}, err
	}

	// A picture with another extension was stored under another key: remove the old object.
	for _, old := range oldPictures {
		if old.Key == profilePic.Key {
			continue
		}
		if err := s.storage.DeleteFile(ctx, old.Key); err != nil {
			s.log.Warn("failed to delete old profile picture", zap.Error(err), zap.String("key", old.Key))
		}
	}

	return s.GetUser(ctx, userID, viewer)
}

// DeleteUser soft-deletes a user. Only root and deu_admin may do it.
func (s *service) DeleteUser(ctx context.Context, userID string, viewer entities.Viewer) error {
	if !viewer.IsGlobalAdmin() {
		return ErrUserForbidden
	}
	return s.repo.DeleteUser(ctx, userID)
}
