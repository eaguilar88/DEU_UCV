package users

import "github.com/eaguilar88/deu/pkg/entities"

type GetUserResponse struct {
	ID             int    `json:"id"`
	CI             string `json:"ci"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Age            int    `json:"age"`
	Gender         string `json:"gender,omitempty"`
	EducationLevel string `json:"education_level,omitempty"`
	Address        string `json:"address,omitempty"`
	Password       string `json:"-"`
	CreatedAt      string `json:"created_at"`
}

type GetUsersResponse struct {
	Users []entities.User    `json:"users,omitempty"`
	Pages entities.PageScope `json:"pages,omitempty"`
}

type CreateUsersResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateUserResponse struct{}

type DeleteUserResponse struct{}
