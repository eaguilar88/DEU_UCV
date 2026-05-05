package users

import (
	"errors"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

const (
	clientDateFormat  = "02-01-2006" // DD-MM-YYYY used on the API surface
	storageDateFormat = "2006-01-02" // YYYY-MM-DD stored in DB
)

func parseDateOfBirth(raw string) (string, error) {
	dob, err := time.Parse(clientDateFormat, raw)
	if err != nil || dob.Format(clientDateFormat) != raw {
		return "", errors.New("fecha_de_nacimiento must be in DD-MM-YYYY format")
	}
	return dob.Format(storageDateFormat), nil
}

// toUserEntity converts CreateUserRequest to a User entity.
func toUserEntity(req CreateUserRequest) (entities.User, error) {
	password, err := generateSecurePassword(req.Password)
	if err != nil {
		return entities.User{}, err
	}

	dob, err := parseDateOfBirth(req.DateOfBirth)
	if err != nil {
		return entities.User{}, err
	}

	return entities.User{
		CI:             req.Document,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dob,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       password,
		Roles: []string{
			entities.RoleNameFromID(entities.RoleVisitante),
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
func toUserUpdateEntity(req UpdateUserRequest) (entities.User, error) {
	dob, err := parseDateOfBirth(req.DateOfBirth)
	if err != nil {
		return entities.User{}, err
	}

	return entities.User{
		ID:             req.ID,
		CI:             req.Document,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		DateOfBirth:    dob,
		Gender:         req.Gender,
		EducationLevel: req.EducationLevel,
		Address:        req.Address,
		Password:       req.Password,
	}, nil
}
