package entities

type CourseRequest struct {
	ID          string
	User        User
	Reviewer    User
	Status      RequestStatus
	Type        string
	Name        string
	Description string
	Comments    string
	ReviewedAt  string
	CreatedAt   string
	UpdatedAtAt string
}
