package entities

type CoursePeriod struct {
	ID              int
	Course          Course
	Participants    []User
	StartDate       string
	EndDate         string
	InscriptionDate string
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       string
}
