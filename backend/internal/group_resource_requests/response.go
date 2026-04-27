package group_resource_requests

import "github.com/eaguilar88/deu/internal/entities"

type GetGroupResourceRequestResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"grupo_id"`
	Type      string `json:"tipo"`
	Content   string `json:"contenido"`
	Status    string `json:"estado"`
	CreatedAt string `json:"creado_en"`
	UpdatedAt string `json:"actualizado_en"`
}

type GetGroupResourceRequestsResponse struct {
	Requests []GetGroupResourceRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope               `json:"paginas"`
}

type CreateGroupResourceRequestResponse struct {
	ID string `json:"id"`
}
