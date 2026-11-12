package users

import "github.com/eaguilar88/deu/internal/entities"

type GetUsersRequest struct {
	PageScope entities.PageScope
}

type GetUserRequest struct {
	ID string
}

type CreateUserRequest struct {
	Document       string `json:"cedula" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	FirstName      string `json:"nombres" validate:"required"`
	LastName       string `json:"apellidos" validate:"required"`
	DateOfBirth    string `json:"fecha_de_nacimiento"`
	Gender         string `json:"genero"`
	EducationLevel string `json:"nivel_educativo"`
	Address        string `json:"direccion"`
	Password       string `json:"password"`
}

type DeleteUserRequest struct {
	ID string `param:"id"`
}

// UpdateUserRequest is a partial update read from multipart/form-data with the same field
// names as CreateUserRequest. A nil field was not sent and is left as it is. The password
// is changed through PUT /auth/password, not here.
type UpdateUserRequest struct {
	ID             string
	Document       *string
	Email          *string
	FirstName      *string
	LastName       *string
	DateOfBirth    *string
	Gender         *string
	EducationLevel *string
	Address        *string
}
