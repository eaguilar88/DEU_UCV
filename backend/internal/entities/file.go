package entities

import (
	"mime/multipart"
)

type OwnerType string

const (
	OwnerTypeProvider       OwnerType = "provider"
	OwnerTypeExtensionGroup OwnerType = "group"
	OwnerTypeCourse         OwnerType = "course"
	OwnerTypeActivity       OwnerType = "group_activity"
	OwnerTypeCourseCycle    OwnerType = "course_cycle"
)

type File struct {
	ID         string
	OwnerID    string
	OwnerType  OwnerType
	Key        string
	Public     bool
	Body       multipart.File
	MetaData   map[string]string
	UploadedBy string
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}
