package courses

import "github.com/eaguilar88/deu/pkg/entities"

type GetCourseResponse struct {
	ID          int                   `json:"id,omitempty"`
	Owner       entities.User         `json:"owner,omitempty"`
	Endorsement entities.Endorsements `json:"endorsement,omitempty"`
	SignedBy    string                `json:"signed_by,omitempty"`
	Objectives  string                `json:"objectives,omitempty"`
	Cost        float64               `json:"cost,omitempty"`
	Content     string                `json:"content,omitempty"`
	Location    string                `json:"location,omitempty"`
	CreatedAt   string                `json:"created_at,omitempty"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
}

type GetCoursesResponse struct {
	Courses []entities.Course  `json:"courses,omitempty"`
	Pages   entities.PageScope `json:"pages,omitempty"`
}

type CreateCoursesResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateCourseResponse struct{}

type DeleteCourseResponse struct{}
