package group_analytics

// AnalyticsFilterRequest recibe los filtros que el frontend envía para calcular las métricas.
type AnalyticsFilterRequest struct {
	CurrentYear int      `json:"anio_actual" query:"anio_actual" validate:"required,min=2000"`
	SinceYear   int      `json:"desde_anio" query:"desde_anio"`
	UntilYear   int      `json:"hasta_anio" query:"hasta_anio"`
	MasterAreas []string `json:"areas_maestras" query:"areas_maestras"`
	GroupID     string   `json:"group_id,omitempty" query:"group_id"`
}
