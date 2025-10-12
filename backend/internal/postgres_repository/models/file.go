package models

import "database/sql"

type File struct {
	ID         string
	OwnerID    string
	OwnerType  string
	FileKey    string
	Public     bool
	Purpose    string
	Metadata   map[string]string
	UploadedBy string
	CreatedAt  string
	DeletedAt  sql.NullString
}

type ProviderFiles struct {
	CI      *File
	RIF     *File
	ISLR    *File
	Resumes []*File
	Others  []*File
}
