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
	IsRenewal bool               `json:"es_renovacion"`
	Approvals []ApprovalResponse `json:"aprobaciones"`
	// Renewal is the group data proposed by a renewal request, only in its detail.
	Renewal *RenewalProposalResponse `json:"renovacion,omitempty"`
}

// RenewalProposalResponse is the group data a renewal will apply once approved. Its fields
// follow the group's response (groups.GetGroupResponse).
type RenewalProposalResponse struct {
	ID                  string                  `json:"id"`
	Name                string                  `json:"nombre"`
	Description         string                  `json:"descripcion,omitempty"`
	Faculty             []string                `json:"facultad"`
	Foundation          string                  `json:"fundacion,omitempty"`
	IsMultidisciplinary bool                    `json:"es_multidisciplinario"`
	Type                []string                `json:"tipo"`
	LogoURL             string                  `json:"imagen_url,omitempty"`
	ProjectURL          string                  `json:"proyecto_url,omitempty"`
	Email               string                  `json:"email,omitempty"`
	Phone               string                  `json:"telefono,omitempty"`
	Objective           string                  `json:"objetivo"`
	Location            string                  `json:"ubicacion,omitempty"`
	Members             []RenewalMemberResponse `json:"miembros"`
	SubmittedAt         string                  `json:"enviado_en"`
}

type RenewalMemberResponse struct {
	Name         string `json:"nombre"`
	CI           int    `json:"cedula"`
	Phone        string `json:"telefono"`
	Email        string `json:"correo"`
	Coordination string `json:"coordinacion"`
	Year         string `json:"año"`
	Faculty      string `json:"facultad"`
	School       string `json:"escuela"`
	DocumentURL  string `json:"documento_url,omitempty"`
	IsLeader     bool   `json:"es_lider"`
	IsActive     bool   `json:"status"`
}

func toRenewalProposal(renewal *entities.GroupRenewal) *RenewalProposalResponse {
	if renewal == nil {
		return nil
	}
	g := renewal.Group
	res := &RenewalProposalResponse{
		ID:                  renewal.ID,
		Name:                g.Name,
		Description:         g.Description,
		Faculty:             make([]string, len(g.Faculty)),
		Foundation:          g.Foundation,
		IsMultidisciplinary: g.IsMultidisciplinary,
		Type:                make([]string, len(g.Type)),
		LogoURL:             fileURL(g.Logo),
		ProjectURL:          fileURL(g.Project),
		Email:               g.Email,
		Phone:               g.Phone,
		Objective:           g.Objective,
		Location:            g.Location,
		Members:             make([]RenewalMemberResponse, len(g.Members)),
		SubmittedAt:         renewal.CreatedAt,
	}
	for i, f := range g.Faculty {
		res.Faculty[i] = string(f)
	}
	for i, t := range g.Type {
		res.Type[i] = string(t)
	}
	for i, m := range g.Members {
		res.Members[i] = RenewalMemberResponse{
			Name:         m.Name,
			CI:           m.CI,
			Phone:        m.Phone,
			Email:        m.Email,
			Coordination: m.Coordination,
			Year:         m.Year,
			Faculty:      string(m.Faculty),
			School:       m.School,
			DocumentURL:  fileURL(m.Document),
			IsLeader:     m.IsLeader,
			IsActive:     m.IsActive,
		}
	}
	return res
}

func fileURL(f *entities.File) string {
	if f == nil {
		return ""
	}
	return f.URL
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
