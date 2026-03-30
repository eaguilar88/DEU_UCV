package entities

type CoursePeriod struct {
	ID              string
	Course          Course
	Participants    []User
	Announcements   []Announcement
	StartDate       string
	EndDate         string
	InscriptionDate string
	IsActive        bool
	ClosedAt        string
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       string
}
