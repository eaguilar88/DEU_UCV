package entities

type CoursePeriod struct {
	ID              string
	Name            string
	Course          Course
	Announcements   []Announcement
	StartDate       string
	EndDate         string
	InscriptionDate string
	Capacity        int
	IsActive        bool
	ClosedAt        string
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       string
}
