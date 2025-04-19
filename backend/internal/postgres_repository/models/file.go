package models

type File struct {
	ID         string
	OwnerID    string
	OwnerType  string
	Path       string
	FileName   string
	Public     bool
	MetaData   map[string]string
	UploadedBy string
	CreatedAt  string
	DeletedAt  string
}
