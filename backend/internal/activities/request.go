package activities

type GetActivityRequest struct {
	ID string `param:"id" validate:"required"`
}

type GetActivitiesRequest struct {
	GroupID               string `query:"group_id"`
	NameSearch            string `query:"name"`
	ActualParticipants    *int   `query:"actual_participants"`
	HasActualParticipants *bool  `query:"has_actual_participants"`
	IsFeatured            *bool  `query:"is_featured"`
	ReportChecked         *bool  `query:"report_checked"`
	StartDate             string `query:"start_date"`
	EndDate               string `query:"end_date"`
	Order                 string `query:"order"`          // "asc" | "desc"
	DisablePaging         bool   `query:"disable_paging"` // true para estadísticas del front
	Page                  int    `query:"page"`
	PerPage               int    `query:"per_page"`
}

type CreateActivityRequest struct {
	GroupID               string `form:"group_id" validate:"required"`
	Name                  string `form:"nombre" validate:"required"`
	Description           string `form:"descripcion"`
	DateStart             string `form:"fecha_inicio"`
	DateEnd               string `form:"fecha_fin"`
	Location              string `form:"ubicacion"`
	KnowledgeArea         string `form:"area_conocimiento"`
	Allies                string `form:"aliados"`
	GroupParticipants     int    `form:"participantes_grupo"`
	EstimatedParticipants int    `form:"participantes_estimados"`
	ActualParticipants    int    `form:"participantes_reales"`
	Financing             string `form:"financiamiento"`
	Comments              string `form:"observaciones"`
	GalleryURL            string `form:"galeria_url"`
	IsFeatured            bool   `form:"destacado"`
}

type UpdateActivityRequest struct {
	ID                    string `param:"id" validate:"required"`
	Name                  string `form:"nombre"`
	Description           string `form:"descripcion"`
	DateStart             string `form:"fecha_inicio"`
	DateEnd               string `form:"fecha_fin"`
	Location              string `form:"ubicacion"`
	KnowledgeArea         string `form:"area_conocimiento"`
	Allies                string `form:"aliados"`
	GroupParticipants     int    `form:"participantes_grupo"`
	EstimatedParticipants int    `form:"participantes_estimados"`
	ActualParticipants    int    `form:"participantes_reales"`
	Financing             string `form:"financiamiento"`
	Comments              string `form:"observaciones"`
	GalleryURL            string `form:"galeria_url"`
	IsFeatured            bool   `form:"destacado"`
}

type ToggleReportCheckRequest struct {
	ID            string `json:"id" validate:"required"`
	ReportChecked bool   `json:"reporte_revisado"`
}

type ToggleFeatureRequest struct {
	ID         string `json:"id" validate:"required"`
	IsFeatured bool   `json:"destacado"`
}

type DeleteActivityRequest struct {
	ID string `param:"id" validate:"required"`
}
