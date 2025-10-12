package courses

import "github.com/eaguilar88/deu/internal/entities"

type GetCoursesRequest struct {
	PageScope entities.PageScope
}

type GetCourseRequest struct {
	ID string `path:"id"`
}

type CreateCourseRequest struct {
	UserID        string
	EndorsementID string  `json:"request_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Objectives    string  `json:"objectives"`
	Content       string  `json:"content"`
	Cost          float64 `json:"cost"`
	Location      string  `json:"location"`
}

type DeleteCourseRequest struct {
	ID string `path:"id"`
}

type UpdateCourseRequest struct {
	ID          string  `path:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Objectives  string  `json:"objectives"`
	Content     string  `json:"content"`
	Cost        float64 `json:"cost"`
	Location    string  `json:"location"`
}
