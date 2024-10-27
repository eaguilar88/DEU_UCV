package entities

type Endorsements struct {
	ID          int               `json:"id,omitempty"`
	User        User              `json:"user,omitempty"`
	Status      EndorsementStatus `json:"status,omitempty"`
	Path        string            `json:"path,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAtAt string            `json:"updated_at_at,omitempty"`
}

type EndorsementStatus string

const (
	ItemStatus_CREATED      EndorsementStatus = "created"
	ItemStatus_APPROVED     EndorsementStatus = "aprobado"
	ItemStatus_REJECTED     EndorsementStatus = "rechazado"
	ItemStatus_UNDER_REVIEW EndorsementStatus = "en revisión"
)
