package users

import (
	"github.com/eaguilar88/deu/pkg/entities"
	"golang.org/x/crypto/bcrypt"
)

func createUserRequestToEntitiesUser(req CreateUserRequest) (entities.User, error) {
	password, err := generateSecurePassword(req.Password)
	if err != nil {
		return entities.User{}, err
	}

	return entities.User{
		CI:             req.Document,
		Username:       req.Username,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       password,
		Roles: []string{
			req.Role,
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

func updateUserRequestToEntitiesUser(req UpdateUserRequest, ID int) entities.User {

	return entities.User{
		ID:             ID,
		CI:             req.Document,
		Username:       req.Username,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    req.DateOfBirth,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       req.Password,
	}
}
