package entities

type CourseCycleCloseRequest struct {
	ID            int64
	CourseCycleID int64
	SubmittedByID string
	Status        RequestStatus
	Comments      string
	ReviewerID    string
	ReviewedAt    string
	CreatedAt     string
	UpdatedAt     string
}
