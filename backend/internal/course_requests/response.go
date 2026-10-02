package course_requests

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type CourseInfo struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"nombre,omitempty"`
	Description string `json:"descripcion,omitempty"`
	Objectives  string `json:"objetivos,omitempty"`
	Duration    string `json:"duracion,omitempty"`
	Content     string `json:"contenido,omitempty"`
	Type        string `json:"tipo,omitempty"`
	Faculty     string `json:"facultad,omitempty"`
	Cost        string `json:"costo,omitempty"`
	Location    string `json:"ubicacion,omitempty"`
	IsActive    bool   `json:"activo,omitempty"`
	// HasDocumentation reports whether a legal contract covers the course.
	HasDocumentation bool   `json:"tiene_documentacion_legal"`
	CreatedAt        string `json:"creado_en,omitempty"`
	UpdatedAt        string `json:"actualizado_en,omitempty"`

	// The rest of the proposal. Only the request detail loads it (list items leave it empty).
	ProviderID        string `json:"id_proveedor,omitempty"`
	OriginFaculty     string `json:"facultad_origen,omitempty"`
	Rationale         string `json:"fundamentacion,omitempty"`
	InstructorProfile string `json:"perfil_docente,omitempty"`
	Profiles          string `json:"perfiles,omitempty"`
	Requirements      string `json:"exigencias,omitempty"`
	Evaluation        string `json:"evaluacion,omitempty"`
	Schedule          string `json:"cronograma,omitempty"`
	Competencies      string `json:"contenido_competencias,omitempty"`
	Bibliography      string `json:"bibliografia,omitempty"`
	Cover             string `json:"portada,omitempty"`
	FacilitatorCV     string `json:"cv_facilitador_url,omitempty"`
}

type GetCourseRequestResponse struct {
	ID         string                 `json:"id,omitempty"`
	Status     entities.RequestStatus `json:"estado,omitempty"`
	Comments   string                 `json:"comentarios,omitempty"`
	ReviewedAt string                 `json:"revisado_en,omitempty"`
	CreatedAt  string                 `json:"creado_en,omitempty"`
	UpdatedAt  string                 `json:"actualizado_en,omitempty"`
	Course     *CourseInfo            `json:"curso,omitempty"`
	// OwnerUserID is the user account of the course's provider.
	OwnerUserID string `json:"usuario_id,omitempty"`

	// Evaluation recorded on approval.
	Score          *float64 `json:"calificacion,omitempty"`
	Classification string   `json:"clasificacion,omitempty"`
	EvaluationFile string   `json:"archivo_evaluacion_url,omitempty"`
}

type GetCourseRequestsResponse struct {
	CourseRequests []GetCourseRequestResponse `json:"solicitudes"`
	Pages          entities.PageScope         `json:"paginas"`
}
