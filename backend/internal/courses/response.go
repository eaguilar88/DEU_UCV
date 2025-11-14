package courses

import (
	"github.com/eaguilar88/deu/internal/course_requests"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

type LatestCoursePeriodInfo struct {
	ID              string `json:"id,omitempty"`
	StartDate       string `json:"fecha_inicio,omitempty"`
	EndDate         string `json:"fecha_fin,omitempty"`
	InscriptionDate string `json:"fecha_inscripcion,omitempty"`
}

type GetCourseResponse struct {
	ID           string                                    `json:"id,omitempty"`
	Content      string                                    `json:"contenido,omitempty"`
	Cost         float64                                   `json:"costo,omitempty"`
	CreatedAt    string                                    `json:"creado_en,omitempty"`
	Description  string                                    `json:"descripcion,omitempty"`
	Duration     int                                       `json:"duracion,omitempty"`
	Endorsement  *course_requests.GetCourseRequestResponse `json:"aval,omitempty"`
	Faculty      string                                    `json:"facultad,omitempty"`
	Location     string                                    `json:"ubicacion,omitempty"`
	Name         string                                    `json:"nombre,omitempty"`
	Objectives   string                                    `json:"objetivos,omitempty"`
	Owner        *users.GetUserResponse                    `json:"propietario,omitempty"`
	Type         string                                    `json:"tipo,omitempty"`
	UpdatedAt    string                                    `json:"actualizado_en,omitempty"`
	LatestPeriod *LatestCoursePeriodInfo                   `json:"ultimo_periodo,omitempty"`
}

type GetCoursesResponse struct {
	Courses []GetCourseResponse `json:"cursos,omitempty"`
	Pages   entities.PageScope  `json:"paginas,omitempty"`
}

type CreateCoursesResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateCourseResponse struct{}

type DeleteCourseResponse struct{}
