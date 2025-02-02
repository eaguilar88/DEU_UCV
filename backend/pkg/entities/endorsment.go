package entities

type Endorsements struct {
	ID          int
	User        User
	Status      EndorsementStatus
	Type        string
	Name        string
	Description string
	Comments    string
	CreatedAt   string
	UpdatedAtAt string
}

type EndorsementStatus string

const (
	ItemStatus_CREATED      EndorsementStatus = "created"
	ItemStatus_APPROVED     EndorsementStatus = "aprobado"
	ItemStatus_REJECTED     EndorsementStatus = "rechazado"
	ItemStatus_UNDER_REVIEW EndorsementStatus = "en revisión"
)
