package group_analytics

// 1. Participantes Reales vs. Estimados por Actividad
type ActivityParticipantsMetric struct {
	Lugar             string `json:"lugar"`             // Nombre de la actividad
	CantidadEsperada  int    `json:"cantidadEsperada"`  // Participantes estimados
	CantidadReal      int    `json:"CantidadReal"`      // Participantes reales
	Integrantes       int    `json:"integrantes"`       // Integrantes organizadores
}

// 2. Volumen de Actividades por Estado
type ActivitiesByStateMetric struct {
	Lugar        string `json:"lugar"`        // Nombre del Estado
	CantidadReal int    `json:"CantidadReal"` // Conteo total de actividades
}

// 3. Total de Participantes Reales vs. Estimados por Grupo
type GroupParticipantsMetric struct {
	Lugar            string `json:"lugar"`            // Nombre del grupo
	CantidadEsperada int    `json:"cantidadEsperada"` // Suma acumulada estimados
	CantidadReal     int    `json:"CantidadReal"`     // Suma acumulada reales
	Integrantes      int    `json:"integrantes"`      // Suma acumulada integrantes
	CantidadDeVeces  int    `json:"cantidadDeVeces"`  // Conteo total de actividades
}

// 4. Cantidad de Actividades por Municipio / Ciudad
type ActivitiesByCityMetric struct {
	Lugar        string `json:"lugar"`        // Nombre de la ciudad / municipio
	CantidadReal int    `json:"CantidadReal"` // Conteo total de actividades
}

// 5. Histórico de Actividades Ejecutadas por Año
type YearlyActivitiesMetric struct {
	Lugar        string `json:"lugar"`        // Año en formato string (ej: "2026")
	CantidadReal int    `json:"CantidadReal"` // Total actividades ejecutadas
}

// GroupAnalyticsResponse concentra la respuesta general de métricas
type GroupAnalyticsResponse struct {
	ActivityParticipants []ActivityParticipantsMetric   `json:"participantes_por_actividad"`
	ActivitiesByState    []ActivitiesByStateMetric      `json:"actividades_por_estado"`
	GroupParticipants    []GroupParticipantsMetric      `json:"participantes_por_grupo"`
	ActivitiesByCity     []ActivitiesByCityMetric       `json:"actividades_por_ciudad"`
	YearlyActivities     []YearlyActivitiesMetric       `json:"historico_por_anio"`
	KnowledgeAreasByYear []map[string]interface{}        `json:"areas_conocimiento_por_anio"`
}
