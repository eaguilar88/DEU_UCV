package entities

type Endorsments struct {
	ID          int              `json:"id,omitempty"`
	User        User             `json:"user,omitempty"`
	Status      EndorsmentStatus `json:"status,omitempty"`
	Path        string           `json:"path,omitempty"`
	CreatedAt   string           `json:"created_at,omitempty"`
	UpdatedAtAt string           `json:"updated_at_at,omitempty"`
}

type EndorsmentStatus string

const (
	ItemStatus_CREATED      EndorsmentStatus = "created"
	ItemStatus_APPROVED     EndorsmentStatus = "aprobado"
	ItemStatus_REJECTED     EndorsmentStatus = "rechazado"
	ItemStatus_UNDER_REVIEW EndorsmentStatus = "en revisión"
)
