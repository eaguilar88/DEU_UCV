package entities

type GroupRequest struct {
	ID         string
	GroupID    string
	GroupName  string
	Faculty    Faculty
	Status     RequestStatus
	Reviewer   *User
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Approvals  []GroupRequest
}

type FacultyPendingCount struct {
	Faculty Faculty `json:"facultad"`
	Count   int     `json:"pendientes"`
}
