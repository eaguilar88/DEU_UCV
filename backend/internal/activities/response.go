package activities

import "github.com/eaguilar88/deu/internal/entities"

type FileResponse struct {
	ID      string `json:"id"`
	Name    string `json:"nombre"`
	URL     string `json:"url"`
	Purpose string `json:"proposito"`
}

type GetActivityResponse struct {
	ID                    string         `json:"id"`
	GroupID               string         `json:"group_id"`
	Name                  string         `json:"nombre"`
	Description           string         `json:"descripcion"`
	Date                  string         `json:"fecha"`
	KnowledgeArea         string         `json:"area_conocimiento"`
	Allies                string         `json:"aliados"`
	EstimatedParticipants int            `json:"participantes_estimados"`
	ActualParticipants    int            `json:"participantes_reales"`
	Financing             string         `json:"financiamiento"`
	Comments              string         `json:"observaciones"`
	Files                 []FileResponse `json:"archivos,omitempty"`
	CreatedAt             string         `json:"creado_en"`
	UpdatedAt             string         `json:"actualizado_en"`
}

type GetActivitiesResponse struct {
	Activities []GetActivityResponse `json:"actividades"`
	PageScope  entities.PageScope    `json:"pagina"`
}

type CreateActivityResponse struct {
	ID string `json:"id"`
}
