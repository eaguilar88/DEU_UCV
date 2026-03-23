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
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

type MailClient interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type service struct {
	repo        Repository
	emailClient MailClient
	log         *zap.Logger
}

func NewService(repository Repository, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repository,
		emailClient: emailClient,
		log:         logger,
	}
}

func (s *service) GetUser(ctx context.Context, userID string) (entities.User, error) {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return entities.User{}, err
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

func (s *service) CreateUser(ctx context.Context, user entities.User) (int64, error) {
    // 1. Validar si ya existe (esto ya lo tienes)
    existingUser, err := s.repo.GetUserByUsername(ctx, user.Email)
    if err != nil && !errors.Is(err, ErrUserNotFound) {
        return -1, fmt.Errorf("failed to check existing user: %w", err)
    }
    if existingUser != nil && existingUser.ID != "" {
        return -1, ErrUserAlreadyExists
    }

    // ✨ EL CAMBIO CRUCIAL:
    // Forzamos el rol de menor rango disponible en tu base de datos.
    // Según tu SQL, el rol para gente común es 'visitante'.
    user.Roles = []string{"visitante"} 

    // 2. Ahora sí, crear el usuario con el rol forzado
    id, err := s.repo.CreateUser(ctx, user)
    if err != nil {
        return -1, fmt.Errorf("error creating new user: %w", err)
    }

    // 3. Envío de correo (lo que te da error 500 si no es correo autorizado)
    if err := s.emailClient.Send(ctx, user.Email, "Welcome to DEU", "Welcome to DEU"); err != nil {
        s.log.Error("welcome email failed", zap.Error(err))
        // OPCIONAL: Podrías quitar el 'return -1' de aquí si quieres que 
        // el usuario se cree aunque falle el email.
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
