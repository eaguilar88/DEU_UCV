package models

import "database/sql"

type Announcement struct {
	ID            string
	CourseCycleID string
	Title         string
	Content       string
	CreatedAt     string
	UpdatedAt     string
	DeletedAt     sql.NullString
}

