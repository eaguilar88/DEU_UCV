package group_requests

import (
	"github.com/eaguilar88/deu/internal/entities"
)

type ApprovalResponse struct {
	ID         string `json:"id"`
	Faculty    string `json:"facultad"`
	Status     string `json:"estado"`
	ReviewedAt string `json:"revisado_en,omitempty"`
}

type GetGroupRequestResponse struct {
	ID        string             `json:"id"`
	GroupID   string             `json:"grupo_id"`
	GroupName string             `json:"grupo_nombre"`
	Comments  string             `json:"comentarios"`
	Status    string             `json:"estado"`
	Faculty   string             `json:"facultad"`
	CreatedAt string             `json:"creado_en"`
	UpdatedAt string             `json:"actualizado_en"`
	Approvals []ApprovalResponse `json:"aprobaciones"`
}

type GetGroupRequestsResponse struct {
	Requests     []GetGroupRequestResponse `json:"solicitudes"`
	Pages        entities.PageScope        `json:"paginas"`
	PendingCount int                       `json:"pendientes_facultad"`
}

type PendingCountsResponse struct {
	Counts []entities.FacultyPendingCount `json:"conteo_pendientes"`
}

func toApprovals(reqs []entities.GroupRequest) []ApprovalResponse {
	out := make([]ApprovalResponse, 0, len(reqs))
	for _, a := range reqs {
		out = append(out, ApprovalResponse{
			ID:         a.ID,
			Faculty:    string(a.Faculty),
			Status:     string(a.Status),
			ReviewedAt: a.ReviewedAt,
		})
	}
	return out
}
