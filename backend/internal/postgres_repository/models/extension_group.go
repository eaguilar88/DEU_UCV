package models

import "database/sql"

type ExtensionGroup struct {
	ID          string
	UserID      string
	Name        string
	Description sql.NullString
	Faculty     string
	Objective   string
	Code        string
	Type        string
	Director    string
	Location    sql.NullString
	IsActive    bool
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
}
