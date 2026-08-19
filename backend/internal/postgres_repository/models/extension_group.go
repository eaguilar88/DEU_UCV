package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type ExtensionGroup struct {
	ID                  string
	UserID              string
	Name                string
	Description         sql.NullString
	Faculty             pq.StringArray
	Foundation          sql.NullString
	IsMultidisciplinary bool
	Objective           string
	Code                string
	Type                pq.StringArray
	Director            string
	Location            sql.NullString
	IsActive            bool
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           sql.NullString
}
