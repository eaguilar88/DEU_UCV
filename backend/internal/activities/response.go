package activities

import "github.com/eaguilar88/deu/internal/entities"

type GetActivityResponse struct {
	ID                    string `json:"id"`
	GroupID               string `json:"group_id"`
	GroupName             string `json:"nombre_grupo,omitempty"`
	Name                  string `json:"nombre"`
	Description           string `json:"descripcion"`
	DateStart             string `json:"fecha_inicio"`
	DateEnd               string `json:"fecha_fin"`
	Location              string `json:"ubicacion"`
	KnowledgeArea         string `json:"area_conocimiento"`
	Allies                string `json:"aliados"`
	GroupParticipants     int    `json:"participantes_grupo"`
	EstimatedParticipants int    `json:"participantes_estimados"`
	ActualParticipants    int    `json:"participantes_reales"`
	Financing             string `json:"financiamiento"`
	Comments              string `json:"observaciones"`
	CoverImage            string `json:"cubierta"`
	ParticipantList       string `json:"lista_participantes,omitempty"`
	GalleryURL            string `json:"galeria_url,omitempty"`
	ReportChecked         bool   `json:"reporte_revisado"`
	IsFeatured            bool   `json:"destacado"`
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
