package entities

const ActivityFileTypeCoverImage = "cubierta"

type ActivityFilter struct {
	GroupID            string
	StartDate          string // YYYY-MM-DD, empty means "no date filter"
	EndDate            string // YYYY-MM-DD
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
	GalleryURL            string
	CoverImage            *File
	CreatedAt             string
	UpdatedAt             string
	DeletedAt             string
}
