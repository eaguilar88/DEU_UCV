package provider_requests

import (
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
)

type GetProviderRequestResponse struct {
	ID         string                 `json:"id"`
	ProviderID string                 `json:"proveedor_id"`
	Status     entities.RequestStatus `json:"estado"`
	Comments   string                 `json:"observaciones,omitempty"`
	ReviewedAt string                 `json:"revisado_en,omitempty"`
	CreatedAt  string                 `json:"creado_en"`
	UpdatedAt  string                 `json:"actualizado_en"`
}

type GetProviderRequestsResponse struct {
	Requests []GetProviderRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope           `json:"paginas"`
}

func providerRequestToResponse(r entities.ProviderRequest) GetProviderRequestResponse {
	return GetProviderRequestResponse{
		ID:         fmt.Sprintf("%d", r.ID),
		ProviderID: fmt.Sprintf("%d", r.ProviderID),
		Status:     r.Status,
		Comments:   r.Comments,
		ReviewedAt: r.ReviewedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
