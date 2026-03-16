package models

import "database/sql"

type Provider struct {
	ID         string
	UserID     string
	Name       sql.NullString
	PartyType  sql.NullString // 'natural' or 'juridical'
	IsInternal sql.NullBool
	Bio        sql.NullString
	Code       sql.NullString
	IsActive   bool
	CreatedAt  sql.NullString
	UpdatedAt  sql.NullString
	DeletedAt  sql.NullString
	UserEmail     string
	UserFirstName string
	UserLastName  string
}
