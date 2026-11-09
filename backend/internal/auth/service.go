package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/jwt"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/eaguilar88/deu/internal/users"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// passwordResetEmailTimeout bounds the background send of the password reset email.
const passwordResetEmailTimeout = time.Minute

var (
	ErrInvalidCurrentPassword = errors.New("invalid current password")
	ErrInvalidResetToken      = errors.New("invalid or expired password reset token")
)

type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUserRoles(ctx context.Context, userID string) ([]entities.UserRole, error)
	GetProviderCodeByUserID(ctx context.Context, userID string) (string, error)
	GetGroupByUserID(ctx context.Context, userID string) (entities.ExtensionGroup, error)
	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetUserCredentialsByID(ctx context.Context, userID string) (*entities.User, error)
	UpdateUserPassword(ctx context.Context, userID, passwordHash string) error
}

type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type service struct {
	repository       Repository
	signer           jwt.Signer
	emailClient      MailClient
	passwordResetURL string
	logger           *zap.Logger
}

func NewService(repo Repository, signer jwt.Signer, emailClient MailClient, passwordResetURL string, logger *zap.Logger) Service {
	return &service{
		repository:       repo,
		signer:           signer,
		emailClient:      emailClient,
		passwordResetURL: passwordResetURL,
		logger:           logger,
	}
}

func (s *service) Login(ctx context.Context, username, password string) (string, *entities.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", nil, err
	}

	userRoles, err := s.repository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}

	roleNames := make([]string, len(userRoles))
	for i, r := range userRoles {
		roleNames[i] = r.Name
	}
	user.Roles = roleNames

	for _, r := range userRoles {
		if r.Faculty != "" {
			user.Faculty = r.Faculty
			break
		}
	}

	providerCode, err := s.repository.GetProviderCodeByUserID(ctx, user.ID)
	if err != nil {
		return "", nil, err
	}
	user.ProviderCode = providerCode

	var groupID, groupName string
	group, err := s.repository.GetGroupByUserID(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, groups.ErrGroupNotFound) {
			return "", nil, err
		}
	} else {
		groupID = group.ID
		groupName = group.Name
	}

	var courseProviderID, courseProviderName string
	provider, err := s.repository.GetProviderByUserID(ctx, user.ID)
	if err != nil {
		if !errors.Is(err, providers.ErrProviderNotFound) {
			return "", nil, err
		}
	} else if provider.IsCourseProvider() {
		courseProviderID = provider.ID
		courseProviderName = provider.Name
	}

	user.GroupID = groupID
	user.GroupName = groupName
	user.CourseProviderID = courseProviderID
	user.CourseProviderName = courseProviderName

	tokenString, err := s.signer.GenerateJWT(user.ID, userRoles, providerCode, groupID, groupName, courseProviderID, courseProviderName)
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

// ChangePassword replaces the user's password after checking the current one.
func (s *service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repository.GetUserCredentialsByID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user credentials", zap.Error(err), zap.String("user_id", userID))
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrInvalidCurrentPassword
	}

	return s.updatePassword(ctx, userID, newPassword)
}

// RequestPasswordReset emails a password reset link when the address belongs to a user. An unknown
// address is not an error, so callers can't tell registered addresses apart. The email is sent in
// the background so the response time doesn't reveal it either.
func (s *service) RequestPasswordReset(ctx context.Context, emailAddr string) error {
	user, err := s.repository.GetUserByUsername(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			s.logger.Info("password reset requested for unknown email")
			return nil
		}
		s.logger.Error("failed to get user for password reset", zap.Error(err))
		return err
	}

	token, err := s.signer.GeneratePasswordResetToken(user.ID, user.Password)
	if err != nil {
		s.logger.Error("failed to generate password reset token", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	resetURL, err := buildResetURL(s.passwordResetURL, token)
	if err != nil {
		s.logger.Error("invalid password reset URL", zap.Error(err))
		return err
	}

	data := email.PasswordResetData{
		Name:     user.FirstName,
		ResetURL: resetURL,
	}
	sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), passwordResetEmailTimeout)
	go func() {
		defer cancel()
		if err := s.emailClient.SendTemplate(sendCtx, user.Email, email.TemplatePasswordReset, data); err != nil {
			s.logger.Error("failed to send password reset email", zap.Error(err), zap.String("user_id", user.ID))
		}
	}()

	return nil
}

// CheckPasswordResetToken reports whether a password reset link token can still be used. It
// returns ErrInvalidResetToken when it can't.
func (s *service) CheckPasswordResetToken(ctx context.Context, token string) error {
	_, err := s.resolveResetToken(ctx, token)
	return err
}

// ResetPassword sets a new password using the token from a password reset link. The token only
// works while the password it was issued for is unchanged.
func (s *service) ResetPassword(ctx context.Context, token, newPassword string) error {
	user, err := s.resolveResetToken(ctx, token)
	if err != nil {
		return err
	}
	return s.updatePassword(ctx, user.ID, newPassword)
}

// resolveResetToken returns the user a password reset token was issued for. A bad, expired or
// already used token, or one whose user no longer exists, is ErrInvalidResetToken.
func (s *service) resolveResetToken(ctx context.Context, token string) (*entities.User, error) {
	userID, fingerprint, err := s.signer.ValidatePasswordResetToken(token)
	if err != nil {
		s.logger.Info("invalid password reset token", zap.Error(err))
		return nil, ErrInvalidResetToken
	}

	user, err := s.repository.GetUserCredentialsByID(ctx, userID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrInvalidResetToken
		}
		s.logger.Error("failed to get user credentials", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	if jwt.PasswordFingerprint(user.Password) != fingerprint {
		s.logger.Info("password reset token already used", zap.String("user_id", userID))
		return nil, ErrInvalidResetToken
	}

	return user, nil
}

func (s *service) updatePassword(ctx context.Context, userID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err), zap.String("user_id", userID))
		return fmt.Errorf("error hashing password: %w", err)
	}

	if err := s.repository.UpdateUserPassword(ctx, userID, string(hash)); err != nil {
		s.logger.Error("failed to update password", zap.Error(err), zap.String("user_id", userID))
		return err
	}
	return nil
}

func buildResetURL(base, token string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("error parsing password reset URL: %w", err)
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
