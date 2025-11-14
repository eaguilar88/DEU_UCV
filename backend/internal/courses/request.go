package courses

import "github.com/eaguilar88/deu/internal/entities"

type GetCoursesRequest struct {
	PageScope entities.PageScope
}

type GetCourseRequest struct {
	ID string `path:"id"`
}

type CreateCourseRequest struct {
	Name        string  `json:"nombre"`
	Description string  `json:"descripcion"`
	Objectives  string  `json:"objetivos"`
	Duration    int     `json:"duracion_horas"`
	Content     string  `json:"contenido"`
	Faculty     string  `json:"facultad"`
	Cost        float64 `json:"costo"`
	Location    string  `json:"ubicacion"`
}

type DeleteCourseRequest struct {
	ID string `path:"id"`
}

type UpdateCourseRequest struct {
	ID          string  `path:"id"`
	Name        string  `form:"nombre"`
	Description string  `form:"descripcion"`
	Objectives  string  `form:"objetivos"`
	Content     string  `form:"contenido"`
	Cost        float64 `form:"costo"`
	Location    string  `form:"ubicacion"`
}
