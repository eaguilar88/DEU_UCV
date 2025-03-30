package entities

type Endorsement struct {
	ID          string
	User        User
	Reviewer    User
	Status      EndorsementStatus
	Type        string
	Name        string
	Description string
	Comments    string
	ReviewedAt  string
	CreatedAt   string
	UpdatedAtAt string
}

type EndorsementStatus string

const (
	EndorsementStatus_CREATED      EndorsementStatus = "created"
	EndorsementStatus_APPROVED     EndorsementStatus = "aprobado"
	EndorsementStatus_REJECTED     EndorsementStatus = "rechazado"
	EndorsementStatus_UNDER_REVIEW EndorsementStatus = "en revisión"
)

type EndorsementType string

const (
	EndorsementType_COURSE EndorsementType = "diplomado"
	EndorsementType_GROUP  EndorsementType = "grupo de extensión"
)
