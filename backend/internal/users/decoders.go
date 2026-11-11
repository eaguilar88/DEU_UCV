package users

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

const (
	clientDateFormat  = "02-01-2006" // DD-MM-YYYY used on the API surface
	storageDateFormat = "2006-01-02" // YYYY-MM-DD stored in DB
)

func parseDateOfBirth(raw string) (string, error) {
	dob, err := time.Parse(clientDateFormat, raw)
	if err != nil {
		return "", fmt.Errorf("error con la fecha de nacimiento: %v", err)
	}

	if dob.Format(clientDateFormat) != raw {
		return "", errors.New("fecha_de_nacimiento debe estar en formato DD-MM-YYYY")
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

// educationLevels are the values of deu.education_level_enum.
var educationLevels = []string{"bachiller", "universitaria_incompleta", "universitaria_completa", "posgrado"}

var fieldValidator = validator.New()

// toUserChanges validates the fields sent in UpdateUserRequest and converts them to a partial
// update. nombres, apellidos, email and cedula can't be cleared; genero and direccion can.
func toUserChanges(req UpdateUserRequest) (entities.UserChanges, error) {
	changes := entities.UserChanges{
		Gender:  req.Gender,
		Address: req.Address,
	}

	required := []struct {
		field string
		value *string
	}{
		{"cedula", req.Document},
		{"email", req.Email},
		{"nombres", req.FirstName},
		{"apellidos", req.LastName},
	}
	for _, r := range required {
		if r.value != nil && strings.TrimSpace(*r.value) == "" {
			return entities.UserChanges{}, fmt.Errorf("%s no puede estar vacío", r.field)
		}
	}

	if req.Document != nil {
		ci, err := strconv.Atoi(strings.TrimSpace(*req.Document))
		if err != nil || ci <= 0 {
			return entities.UserChanges{}, errors.New("cedula debe ser un número positivo")
		}
		document := strconv.Itoa(ci)
		changes.CI = &document
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if err := fieldValidator.Var(email, "email"); err != nil {
			return entities.UserChanges{}, errors.New("email no es válido")
		}
		changes.Email = &email
	}

	changes.FirstName = req.FirstName
	changes.LastName = req.LastName

	if req.DateOfBirth != nil {
		dob, err := parseDateOfBirth(*req.DateOfBirth)
		if err != nil {
			return entities.UserChanges{}, err
		}
		changes.DateOfBirth = &dob
	}

	if req.EducationLevel != nil {
		if !slices.Contains(educationLevels, *req.EducationLevel) {
			return entities.UserChanges{}, fmt.Errorf("nivel_educativo debe ser uno de: %s", strings.Join(educationLevels, ", "))
		}
		changes.EducationLevel = req.EducationLevel
	}

	return changes, nil
}
