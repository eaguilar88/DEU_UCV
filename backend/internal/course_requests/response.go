package course_requests

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type CourseInfo struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"nombre,omitempty"`
	Description string `json:"descripcion,omitempty"`
	Objectives  string `json:"objetivos,omitempty"`
	Duration    int    `json:"duracion,omitempty"`
	Content     string `json:"contenido,omitempty"`
	Type        string `json:"tipo,omitempty"`
	Faculty     string `json:"facultad,omitempty"`
	Cost        string `json:"costo,omitempty"`
	Location    string `json:"ubicacion,omitempty"`
	IsActive    bool   `json:"activo,omitempty"`
	CreatedAt   string `json:"creado_en,omitempty"`
	UpdatedAt   string `json:"actualizado_en,omitempty"`
}

type GetCourseRequestResponse struct {
	ID          string                 `json:"id,omitempty"`
	Status      entities.RequestStatus `json:"estado,omitempty"`
	Comments    string                 `json:"comentarios,omitempty"`
	ReviewedAt  string                 `json:"revisado_en,omitempty"`
	CreatedAt   string                 `json:"creado_en,omitempty"`
	UpdatedAtAt string                 `json:"actualizado_en,omitempty"`
	Course      *CourseInfo            `json:"curso,omitempty"`
}

type GetCourseRequestsResponse struct {
	CourseRequests []GetCourseRequestResponse `json:"solicitudes"`
	Pages          entities.PageScope         `json:"paginas"`
}
