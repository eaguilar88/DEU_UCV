package users

import "github.com/eaguilar88/deu/pkg/entities"

type GetUsersRequest struct {
	PageScope entities.PageScope
}

type GetUserRequest struct {
	UserID string
}

type CreateUserRequest struct {
	Username string `json:"username"`
	// FirstName      string `json:"first_name"`
	// LastName       string `json:"last_name"`
	// DateOfBirth    string `json:"date_of_birth"`
	// Gender         string `json:"gender"`
	// EducationLevel string `json:"education_level"`
	// Address        string `json:"address"`
	Password string `json:"password"`
}
