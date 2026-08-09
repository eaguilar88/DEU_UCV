package models

import "database/sql"

type ExtensionGroup struct {
	ID                  string
	UserID              string
	Name                string
	Description         sql.NullString
	Faculty             string
	Foundation          sql.NullString
	IsMultidisciplinary bool
	Objective           string
	Code                string
	Type                string
	Director            string
	Location            sql.NullString
	IsActive            bool
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           sql.NullString
}
