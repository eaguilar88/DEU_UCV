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
	Course        CloseRequestCourse     `json:"curso"`
	CohortName    string                 `json:"nombre_cohorte,omitempty"`
	// Files is only set on the request detail.
	Files *CloseRequestFiles `json:"archivos,omitempty"`
}

type CloseRequestCourse struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}

// CloseRequestFiles holds short-lived pre-signed URLs of the evidence files.
type CloseRequestFiles struct {
	Participants string `json:"participantes_url,omitempty"`
	Vouchers     string `json:"vouchers_url,omitempty"`
	Survey       string `json:"encuesta_url,omitempty"`
}

type GetCloseRequestsResponse struct {
	Requests []GetCloseRequestResponse `json:"solicitudes"`
	Pages    entities.PageScope        `json:"paginas"`
}

func closeRequestToResponse(r entities.CourseCycleCloseRequest) GetCloseRequestResponse {
	resp := GetCloseRequestResponse{
		ID:            fmt.Sprintf("%d", r.ID),
		CourseCycleID: fmt.Sprintf("%d", r.CourseCycleID),
		SubmittedByID: r.SubmittedByID,
		Status:        r.Status,
		Comments:      r.Comments,
		ReviewedAt:    r.ReviewedAt,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Course:        CloseRequestCourse{ID: r.CourseID, Name: r.CourseName},
		CohortName:    r.CohortName,
	}
	if r.ParticipantsFile != nil || r.VouchersFile != nil || r.SurveyFile != nil {
		resp.Files = &CloseRequestFiles{
			Participants: fileURL(r.ParticipantsFile),
			Vouchers:     fileURL(r.VouchersFile),
			Survey:       fileURL(r.SurveyFile),
		}
	}
	return resp
}

func fileURL(f *entities.File) string {
	if f == nil {
		return ""
	}
	return f.URL
}
