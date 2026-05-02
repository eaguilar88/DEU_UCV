package entities

type GroupRequest struct {
	ID         string
	GroupID    string
	Faculty    Faculty
	Status     RequestStatus
	Reviewer   *User
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Approvals  []GroupRequest
}
