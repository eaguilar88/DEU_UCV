package course_requests

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type GetCourseRequestResponse struct {
	ID          string                 `json:"id,omitempty"`
	User        *users.GetUserResponse `json:"user,omitempty"`
	Reviewer    *users.GetUserResponse `json:"reviewer,omitempty"`
	Status      entities.RequestStatus `json:"status,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Comments    string                 `json:"comments,omitempty"`
	ReviewedAt  string                 `json:"reviewed_at,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
	UpdatedAtAt string                 `json:"updated_at_at,omitempty"`
}

type GetCourseRequestsResponse struct {
	CourseRequests []GetCourseRequestResponse `json:"requests"`
	Pages          entities.PageScope         `json:"pages"`
}

type CreateCourseRequestResponse struct {
	ID string `json:"id,omitempty"`
}
