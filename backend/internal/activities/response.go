package activities

import "github.com/eaguilar88/deu/internal/entities"

type GetActivityResponse struct {
	ID                    string `json:"id"`
	GroupID               string `json:"group_id"`
	Name                  string `json:"nombre"`
	Description           string `json:"descripcion"`
	Date                  string `json:"fecha"`
	KnowledgeArea         string `json:"area_conocimiento"`
	Allies                string `json:"aliados"`
	EstimatedParticipants int    `json:"participantes_estimados"`
	ActualParticipants    int    `json:"participantes_reales"`
	Financing             string `json:"financiamiento"`
	Comments              string `json:"observaciones"`
	CoverImage            string `json:"cubierta,omitempty"`
	GalleryURL            string `json:"gallery_url,omitempty"`
	CreatedAt             string `json:"creado_en"`
	UpdatedAt             string `json:"actualizado_en"`
}

type GetActivitiesResponse struct {
	Activities []GetActivityResponse `json:"actividades"`
	PageScope  entities.PageScope    `json:"pagina"`
}

type CreateActivityResponse struct {
	ID string `json:"id"`
}
