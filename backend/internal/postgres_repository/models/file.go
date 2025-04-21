package models

type File struct {
	ID         string
	OwnerID    string
	OwnerType  string
	FileKey    string
	Public     bool
	Metadata   map[string]string
	UploadedBy string
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}
