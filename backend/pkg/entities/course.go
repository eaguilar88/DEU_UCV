package entities

type Course struct {
	ID          int
	Owner       User
	Endorsement Endorsements
	Name        string
	Description string
	Objectives  string
	Cost        float64
	Content     string
	Location    string
	CreatedAt   string
	UpdatedAt   string
}
