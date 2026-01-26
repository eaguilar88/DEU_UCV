package models

import "database/sql"

type Course struct {
	ID          string
	Name        string
	Description sql.NullString
	OwnerID     string
	Content     string
	Objectives  sql.NullString
	Duration    sql.NullInt64
	Type        sql.NullString
	Faculty     sql.NullString
	Cost        sql.NullString
	Location    sql.NullString
	IsActive    bool
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   sql.NullString
}
