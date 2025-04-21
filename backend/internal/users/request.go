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
	Username       string `json:"username"        validate:"required,email"`
	FirstName      string `json:"first_name"      validate:"required"`
	LastName       string `json:"last_name"       validate:"required"`
	Role           string `json:"role"            validate:"required"`
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender"`
	EducationLevel string `json:"education_level"`
	Address        string `json:"address"`
	Password       string `json:"password"`
}

type DeleteUserRequest struct {
	ID string `param:"id"`
}

type UpdateUserRequest struct {
	ID             string `param:"id" validate:"required"`
	Document       string `                               json:"ci"`
	FirstName      string `                               json:"first_name"`
	LastName       string `                               json:"last_name"`
	DateOfBirth    string `                               json:"date_of_birth"`
	Gender         string `                               json:"gender"`
	EducationLevel string `                               json:"education_level"`
	Address        string `                               json:"address"`
	Password       string `                               json:"password"`
}
