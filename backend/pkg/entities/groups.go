package entities

type ExtensionGroup struct {
	ID          string
	Name        string
	Description string
	Owner       User
	Endorsement Endorsement
	Objective   string
	Location    string
	Active      bool
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
}
