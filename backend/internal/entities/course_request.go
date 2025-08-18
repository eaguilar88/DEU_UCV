package entities

type CourseRequest struct {
	ID          string
	User        User
	Reviewer    User
	Status      CourseRequestStatus
	Type        string
	Name        string
	Description string
	Comments    string
	ReviewedAt  string
	CreatedAt   string
	UpdatedAtAt string
}

type CourseRequestStatus string

const (
	CourseRequestStatus_CREATED      CourseRequestStatus = "created"
	CourseRequestStatus_APPROVED     CourseRequestStatus = "aprobado"
	CourseRequestStatus_REJECTED     CourseRequestStatus = "rechazado"
	CourseRequestStatus_UNDER_REVIEW CourseRequestStatus = "en revisión"
)

type CourseRequestType string

const (
	CourseRequestType_COURSE CourseRequestType = "diplomado"
	CourseRequestType_GROUP  CourseRequestType = "grupo de extensión"
)
