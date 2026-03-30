package course_cycle_close_requests

import (
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
)

type GetCloseRequestResponse struct {
	ID            string                 `json:"id"`
	CourseCycleID string                 `json:"course_cycle_id"`
	SubmittedByID string                 `json:"enviado_por_id"`
	Status        entities.RequestStatus `json:"estado"`
	Comments      string                 `json:"observaciones,omitempty"`
	ReviewedAt    string                 `json:"revisado_en,omitempty"`
	CreatedAt     string                 `json:"creado_en"`
	UpdatedAt     string                 `json:"actualizado_en"`
}

type GetCloseRequestsResponse struct {
	Requests []GetCloseRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope        `json:"paginas"`
}

func closeRequestToResponse(r entities.CourseCycleCloseRequest) GetCloseRequestResponse {
	return GetCloseRequestResponse{
		ID:            fmt.Sprintf("%d", r.ID),
		CourseCycleID: fmt.Sprintf("%d", r.CourseCycleID),
		SubmittedByID: r.SubmittedByID,
		Status:        r.Status,
		Comments:      r.Comments,
		ReviewedAt:    r.ReviewedAt,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
