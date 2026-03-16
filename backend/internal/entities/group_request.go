package entities

type GroupRequest struct {
	ID        string
	GroupID   string
	Faculty   Faculty
	Status    RequestStatus
	Reviewer  *User
	Comments  string
	CreatedAt string
	UpdatedAt string
}
