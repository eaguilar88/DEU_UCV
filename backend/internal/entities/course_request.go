package entities

type CourseRequest struct {
	ID         string
	User       User
	Reviewer   User
	Status     RequestStatus
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Course     *Course // Associated course information
}
