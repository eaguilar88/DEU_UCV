package entities

type Endorsments struct {
	User        User
	Status      EndorsmentStatus
	Path        string
	CreatedAt   string
	UpdatedAtAt string
}

type EndorsmentStatus int

const (
	ItemStatus_CREATED EndorsmentStatus = iota
	ItemStatus_APPROVED
	ItemStatus_REJECTED
	ItemStatus_UNDER_REVIEW
)
