package entities

const (
	ActivityFileTypeCoverImage = "cubierta"
	ActivityFileTypeListParticipants = "lista_participantes"
)

type ActivityFilter struct {
	GroupID               string
	NameSearch            string // Filtro por coincidencia de nombre
	StartDate             string // YYYY-MM-DD, empty means "no date filter"
	EndDate               string // YYYY-MM-DD
	Deleted               bool
	ActualParticipants    *int
	HasActualParticipants *bool
	IsFeatured            *bool 
	ReportChecked         *bool
	Order                 string // "asc" o "desc"
	DisablePaging         bool   // Si es true, omite el LIMIT / OFFSET para estadísticas
}

type Activity struct {
	ID                    string
	GroupID               string
	GroupName             string `json:",omitempty"`
	Name                  string
	Description           string
	DateStart             string
	DateEnd               string
	Location              string
	KnowledgeArea         string
	Allies                string
	GroupParticipants     int
	EstimatedParticipants int
	ActualParticipants    int
	Financing             string
	Comments              string
	GalleryURL            string
	ReportChecked         bool
	IsFeatured            bool
	CoverImage            *File
	ParticipantList       *File
	CreatedAt             string
	UpdatedAt             string
	DeletedAt             string
}

type CurrentActivitySummary struct {
	ID   string `json:"id"`
	Name string `json:"nombre"`
}

type GroupDashboardSummary struct {
	PlannedCount       int                      `json:"actividades_planificadas"`
	PendingReportCount int                      `json:"reportes_pendientes"`
	CurrentActivities  []CurrentActivitySummary `json:"actividades_actuales"`
}
