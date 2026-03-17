package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type LatestCoursePeriodInfo struct {
	ID              string `json:"id,omitempty"`
	StartDate       string `json:"fecha_inicio,omitempty"`
	EndDate         string `json:"fecha_fin,omitempty"`
	InscriptionDate string `json:"fecha_inscripcion,omitempty"`
}

type GetCourseResponse struct {
	ID                string                  `json:"id,omitempty"`
	Name              string                  `json:"nombre,omitempty"`
	Description       string                  `json:"descripcion,omitempty"`
	Cover             string                  `json:"portada,omitempty"`
	Objectives        string                  `json:"objetivos,omitempty"`
	Rationale         string                  `json:"fundamentacion,omitempty"`
	Duration          string                  `json:"duracion,omitempty"`
	Cost              string                  `json:"estructura_costos,omitempty"`
	InstructorProfile string                  `json:"perfil_docente,omitempty"`
	Profiles          string                  `json:"perfiles,omitempty"`
	Requirements      string                  `json:"exigencias,omitempty"`
	Content           string                  `json:"estructura_curricular,omitempty"`
	Evaluation        string                  `json:"evaluacion,omitempty"`
	Schedule          string                  `json:"cronograma,omitempty"`
	ProviderCode      string                  `json:"codigo_proveedor,omitempty"`
	ProviderID        string                  `json:"id_proveedor,omitempty"`
	Faculty           string                  `json:"facultad,omitempty"`
	Location          string                  `json:"ubicacion,omitempty"`
	Type              string                  `json:"tipo,omitempty"`
	CreatedAt         string                  `json:"creado_en,omitempty"`
	UpdatedAt         string                  `json:"actualizado_en,omitempty"`
	LatestPeriod      *LatestCoursePeriodInfo `json:"ultimo_periodo,omitempty"`
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
