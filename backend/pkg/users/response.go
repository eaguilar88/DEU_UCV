package users

import "github.com/eaguilar88/deu/pkg/entities"

type GetUserResponse struct {
	ID             int    `json:"id,omitempty"`
	CI             string `json:"ci,omitempty"`
	Username       string `json:"username,omitempty"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	DateOfBirth    string `json:"date_of_birth,omitempty"`
	Age            int    `json:"age,omitempty"`
	Gender         string `json:"gender,omitempty"`
	EducationLevel string `json:"education_level,omitempty"`
	Code           string `json:"code,omitempty"`
	Address        string `json:"address,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}

type GetUsersResponse struct {
	Users []GetUserResponse  `json:"users"`
	Pages entities.PageScope `json:"pages"`
}

type CreateUsersResponse struct {
	ID string `json:"id,"`
}
