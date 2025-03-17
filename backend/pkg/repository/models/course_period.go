package models

type CoursePeriod struct {
	ID              int
	Course          Course
	StartDate       string
	EndDate         string
	InscriptionDate string
	CreatedAt       string
	UpdatedAt       string
	DeletedAt       string
}
