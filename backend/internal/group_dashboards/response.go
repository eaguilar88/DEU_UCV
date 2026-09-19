package group_dashboards

type GroupDashboardResponse struct {
	GroupID            string `json:"group_id"`
	UpcomingActivities int    `json:"actividades_futuras"`
	PendingReports     int    `json:"reportes_pendientes"`
}

type FacultyDashboardResponse struct {
	Faculty         string `json:"facultad"`
	PendingRequests int    `json:"solicitudes_pendientes"`
	TotalGroups     int    `json:"grupos_totales"`
}

type ResourceRequestsByFaculty struct {
	Faculty string `json:"facultad"`
	Count   int    `json:"solicitudes"`
}

type DeuDashboardResponse struct {
	PendingRequests           int                         `json:"solicitudes_pendientes_deu"`
	ResourceRequestsByFaculty []ResourceRequestsByFaculty `json:"solicitudes_recursos_por_facultad"`
	ActiveGroupsCount         int                         `json:"grupos_activos_totales"`
	InactiveGroupsCount       int                         `json:"grupos_inactivos_totales"`
}
