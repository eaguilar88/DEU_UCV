package entities

const ActivityFileTypeReport = "reporte"

type ActivityFilter struct {
	GroupID            string
	Date               string
	Deleted            bool
	ActualParticipants *int
}

type Activity struct {
	ID                    string
	GroupID               string
	Name                  string
	Description           string
	Date                  string
	KnowledgeArea         string
	Allies                string
	EstimatedParticipants int
	ActualParticipants    int
	Financing             string
	Comments              string
	Files                 GroupedFiles
	CreatedAt             string
	UpdatedAt             string
	DeletedAt             string
}
