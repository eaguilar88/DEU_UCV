package entities

type CoursePeriod struct {
	ID              int
	Course          Course
	Participants    []User
	StartDate       string
	EndDate         string
	InscriptionDate string
	IsActive        bool
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       string
}
