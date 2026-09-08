package models

import "database/sql"

type GroupMember struct {
	ID           string
	GroupID      string
	Name         string
	CI           int
	Phone        sql.NullString
	Email        sql.NullString
	Coordination sql.NullString
	Year         sql.NullString
	Faculty      string
	School       sql.NullString
	IsLeader     bool
	IsActive     bool
}
