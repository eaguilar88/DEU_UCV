package users

import (
	"github.com/eaguilar88/deu/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

// toUserEntity converts CreateUserRequest to a User entity.
func toUserEntity(req CreateUserRequest) (entities.User, error) {
	password, err := generateSecurePassword(req.Password)
	if err != nil {
		return entities.User{}, err
	}

	return entities.User{
		CI:             req.Document,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       password,
		Roles: []string{
			entities.RoleNameFromID(entities.RoleCoordinador),
		},
	}, nil
}

func generateSecurePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// toUserUpdateEntity converts UpdateUserRequest to a User entity.
func toUserUpdateEntity(req UpdateUserRequest) entities.User {
	return entities.User{
		ID:             req.ID,
		CI:             req.Document,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       req.Password,
	}
}
