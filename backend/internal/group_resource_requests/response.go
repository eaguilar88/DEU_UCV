package group_resource_requests

import "github.com/eaguilar88/deu/internal/entities"

type GetGroupResourceRequestResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"grupo_id"`
	GroupName string `json:"grupo_nombre"`
	Type      string `json:"tipo"`
	Content   string `json:"contenido"`
	Status    string `json:"estado"`
	CreatedAt string `json:"creado_en"`
	UpdatedAt string `json:"actualizado_en"`
}

type GetGroupResourceRequestsResponse struct {
	Requests []GetGroupResourceRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope                `json:"paginas"`
}

type GetGroupResourceRequestsByFacultyResponse struct {
	Requests     []GetGroupResourceRequestResponse `json:"solicitudes"`
	PendingCount int                               `json:"solicitudes_pendientes"`
	Pages        entities.PageScope                `json:"paginas"`
}

type FacultyPendingCount struct {
	Faculty      string `json:"facultad"`
	PendingCount int    `json:"solicitudes_pendientes"`
}

type CreateGroupResourceRequestResponse struct {
	ID string `json:"id"`
}
