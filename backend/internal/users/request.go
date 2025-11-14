package users

import "github.com/eaguilar88/deu/internal/entities"

type GetUsersRequest struct {
	PageScope entities.PageScope
}

type GetUserRequest struct {
	ID string
}

type CreateUserRequest struct {
	Document       string `json:"ci"`
	Username       string `json:"nombre_usuario" validate:"required,email"`
	FirstName      string `json:"primer_nombre" validate:"required"`
	LastName       string `json:"apellido" validate:"required"`
	Role           string `json:"rol" validate:"required"`
	DateOfBirth    string `json:"fecha_nacimiento"`
	Gender         string `json:"genero"`
	EducationLevel string `json:"nivel_educacion"`
	Address        string `json:"direccion"`
	Password       string `json:"password"`
}

type DeleteUserRequest struct {
	ID string `param:"id"`
}

type UpdateUserRequest struct {
	ID             string `param:"id" validate:"required"`
	Document       string `json:"ci"`
	FirstName      string `json:"primer_nombre"`
	LastName       string `json:"apellido"`
	DateOfBirth    string `json:"fecha_nacimiento"`
	Gender         string `json:"genero"`
	EducationLevel string `json:"nivel_educacion"`
	Address        string `json:"direccion"`
	Password       string `json:"password"`
}
