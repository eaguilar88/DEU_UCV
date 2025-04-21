package models

type Provider struct {
	ID        string
	UserID    string
	FirstName string
	LastName  string
	Code      string
	Active    bool
	Files     []*File
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}
