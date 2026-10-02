package courses

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type LatestCoursePeriodInfo struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"nombre_cohorte,omitempty"`
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
	OriginFaculty     string                  `json:"facultad_origen,omitempty"`
	Location          string                  `json:"ubicacion,omitempty"`
	Type              string                  `json:"tipo,omitempty"`
	CreatedAt         string                  `json:"creado_en,omitempty"`
	UpdatedAt         string                  `json:"actualizado_en,omitempty"`
	LatestPeriod      *LatestCoursePeriodInfo `json:"ultimo_periodo,omitempty"`
	// OwnerUserID is the user account of the course's provider (only on the course detail).
	OwnerUserID      string              `json:"usuario_id,omitempty"`
	IsActive         bool                `json:"activo"`
	HasDocumentation bool                `json:"tiene_documentacion_legal"`
	ManagementStatus string              `json:"estado_gestion,omitempty"`
	Competencies     string              `json:"contenido_competencias,omitempty"`
	Bibliography     string              `json:"bibliografia,omitempty"`
	FacilitatorCV    string              `json:"cv_facilitador_url,omitempty"`
	Provider         *CourseProviderInfo `json:"proveedor,omitempty"`
}

type CourseProviderInfo struct {
	Name    string `json:"nombre,omitempty"`
	LogoURL string `json:"logo_url,omitempty"`
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
