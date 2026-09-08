package entities

import "io"

type OwnerType string

const (
	OwnerTypeProvider       OwnerType = "provider"
	OwnerTypeExtensionGroup OwnerType = "group"
	OwnerTypeCourse         OwnerType = "course"
	OwnerTypeActivity       OwnerType = "group_activity"
	OwnerTypeCourseCycle    OwnerType = "course_cycle"
	OwnerTypeUser           OwnerType = "user"
	OwnerTypeGroupMember    OwnerType = "group_member"
)

func (ot OwnerType) String() string {
	return string(ot)
}

type File struct {
	ID         string
	OwnerID    string
	OwnerType  OwnerType
	Name       string
	Key        string
	URL        string
	Public     bool
	Purpose    string
	Version    int
	Body       io.Reader
	MetaData   map[string]string
	UploadedBy string
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}

type GroupedFiles map[string][]*File

func (gf GroupedFiles) GetSingleFile(purpose string) *File {
	if files, ok := gf[purpose]; ok && len(files) > 0 {
		return files[0]
	}
	return nil
}

func (gf GroupedFiles) GetMultipleFiles(purpose string) []*File {
	if files, ok := gf[purpose]; ok && len(files) > 0 {
		return files
	}
	return nil
}

func (gf GroupedFiles) GetAllFiles() []*File {
	var allFiles []*File
	for _, files := range gf {
		allFiles = append(allFiles, files...)
	}
	return allFiles
}
