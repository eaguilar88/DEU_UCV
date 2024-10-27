package courses

import "github.com/eaguilar88/deu/pkg/entities"

type GetCoursesRequest struct {
	PageScope entities.PageScope
}

type GetCourseRequest struct {
	ID string
}

type CreateCourseRequest struct {
	Document       string `json:"ci"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender"`
	EducationLevel string `json:"education_level"`
	Address        string `json:"address"`
	Password       string `json:"password"`
}

type DeleteCourseRequest struct {
	ID string
}

type UpdateCourseRequest struct {
	ID             string
	Document       string `json:"ci"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Gender         string `json:"gender"`
	EducationLevel string `json:"education_level"`
	Address        string `json:"address"`
	Password       string `json:"password"`
}
