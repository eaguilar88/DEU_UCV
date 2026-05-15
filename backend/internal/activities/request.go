package activities

type GetActivityRequest struct {
	ID string `param:"id" validate:"required"`
}

type GetActivitiesRequest struct {
	GroupID            string `query:"group_id" validate:"required"`
	Date               string `query:"date"`
	ActualParticipants *int   `query:"actual_participants"`
	Page               int    `query:"page" validate:"required"`
	PerPage            int    `query:"per_page" validate:"required"`
}

type CreateActivityRequest struct {
	GroupID               string `form:"group_id" validate:"required"`
	Name                  string `form:"nombre" validate:"required"`
	Description           string `form:"descripcion"`
	Date                  string `form:"fecha"`
	KnowledgeArea         string `form:"area_conocimiento"`
	Allies                string `form:"aliados"`
	EstimatedParticipants int    `form:"participantes_estimados"`
	ActualParticipants    int    `form:"participantes_reales"`
	Financing             string `form:"financiamiento"`
	Comments              string `form:"observaciones"`
}

type UpdateActivityRequest struct {
	ID                    string `param:"id" validate:"required"`
	Name                  string `form:"nombre"`
	Description           string `form:"descripcion"`
	Date                  string `form:"fecha"`
	KnowledgeArea         string `form:"area_conocimiento"`
	Allies                string `form:"aliados"`
	EstimatedParticipants int    `form:"participantes_estimados"`
	ActualParticipants    int    `form:"participantes_reales"`
	Financing             string `form:"financiamiento"`
	Comments              string `form:"observaciones"`
}

type DeleteActivityRequest struct {
	ID string `param:"id" validate:"required"`
}
