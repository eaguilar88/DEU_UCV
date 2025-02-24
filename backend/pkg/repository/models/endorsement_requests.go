package models

import "database/sql"

type Endorsement struct {
	ID          int
	UserID      int
	Type        string
	Name        sql.NullString
	Description sql.NullString
	Status      string
	Comments    sql.NullString
	CreatedAt   string
	UpdatedAt   string
}
