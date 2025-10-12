package course_requests

import "github.com/eaguilar88/deu/internal/entities"

type GetCourseRequestsRequest struct {
	PageScope entities.PageScope
}

type GetCourseRequestRequest struct {
	ID string
}

type CreateCourseRequestRequest struct {
	UserID      string `json:"user_id,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Comments    string `json:"comments,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAtAt string `json:"updated_at_at,omitempty"`
}

type DeleteCourseRequestRequest struct {
	ID string `path:"id"`
}

type UpdateCourseRequestRequest struct {
	ID          string `path:"id"`
	UserID      string `json:"user_id,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Comments    string `json:"comments,omitempty"`
}
