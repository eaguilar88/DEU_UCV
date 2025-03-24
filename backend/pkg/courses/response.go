package courses

import (
	"github.com/eaguilar88/deu/pkg/endorsements"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
)

type GetCourseResponse struct {
	ID          int                                  `json:"id,omitempty"`
	Content     string                               `json:"content,omitempty"`
	Cost        float64                              `json:"cost,omitempty"`
	CreatedAt   string                               `json:"created_at,omitempty"`
	Description string                               `json:"description,omitempty"`
	Endorsement *endorsements.GetEndorsementResponse `json:"endorsement,omitempty"`
	Location    string                               `json:"location,omitempty"`
	Name        string                               `json:"name,omitempty"`
	Objectives  string                               `json:"objectives,omitempty"`
	Owner       *users.GetUserResponse               `json:"owner,omitempty"`
	UpdatedAt   string                               `json:"updated_at,omitempty"`
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
