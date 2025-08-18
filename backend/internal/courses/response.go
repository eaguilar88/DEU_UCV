package courses

import (
	"github.com/eaguilar88/deu/internal/course_request"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetCourseResponse struct {
	ID          string                                   `json:"id,omitempty"`
	Content     string                                   `json:"content,omitempty"`
	Cost        float64                                  `json:"cost,omitempty"`
	CreatedAt   string                                   `json:"created_at,omitempty"`
	Description string                                   `json:"description,omitempty"`
	Endorsement *course_request.GetCourseRequestResponse `json:"endorsement,omitempty"`
	Location    string                                   `json:"location,omitempty"`
	Name        string                                   `json:"name,omitempty"`
	Objectives  string                                   `json:"objectives,omitempty"`
	Owner       *users.GetUserResponse                   `json:"owner,omitempty"`
	UpdatedAt   string                                   `json:"updated_at,omitempty"`
}

type GetCoursesResponse struct {
	Courses []GetCourseResponse `json:"courses,omitempty"`
	Pages   entities.PageScope  `json:"pages,omitempty"`
}

type CreateCoursesResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateCourseResponse struct{}

type DeleteCourseResponse struct{}
