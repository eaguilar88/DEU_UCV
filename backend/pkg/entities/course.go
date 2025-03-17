package entities

type Course struct {
	ID          int
	Content     string
	Cost        float64
	Description string
	Endorsement Endorsements
	Location    string
	Name        string
	Objectives  string
	Owner       User
	Periods     []CoursePeriod
	CreatedAt   string
	UpdatedAt   string
}
