package group_requests

import "github.com/eaguilar88/deu/internal/entities"

type GetGroupRequestResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"grupo_id"`
	Comments  string `json:"comentarios"`
	Status    string `json:"estado"`
	CreatedAt string `json:"creado_en"`
	UpdatedAt string `json:"actualizado_en"`
}

type GetGroupRequestsResponse struct {
	Requests []GetGroupRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope        `json:"paginas"`
}
