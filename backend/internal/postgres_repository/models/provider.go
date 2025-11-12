package models

import "database/sql"

type Provider struct {
	ID        string
	UserID    string
	FirstName string
	LastName  string
	Code      string
	Active    bool
	Files     []*File
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
	DeletedAt sql.NullString
}
