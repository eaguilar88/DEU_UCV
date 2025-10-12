package entities

type GroupAuthRequest struct {
	ID        string
	GroupID   string
	Faculty   Faculty
	Status    RequestStatus
	Reviewer  *User
	Comments  string
	CreatedAt string
	UpdatedAt string
}
