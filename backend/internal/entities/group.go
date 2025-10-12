package entities

type GroupType string

const (
	// Group types
	CulturalGroupType = "cultural"
	SportsGroupType   = "sports"
	AcademicGroupType = "academic"

	// File types for groups
	GroupFileTypeLogo          = "logo"
	GroupFileTypeFinancingPlan = "financing_plan"
	GroupFileTypeGroupProject  = "group_project"
)

type ExtensionGroup struct {
	ID          string
	Name        string
	Owner       *User
	LeadName    string
	Description string
	Objective   string
	Location    string
	Type        GroupType
	Faculty     Faculty
	Files       *GroupFiles
	Members     []GroupMember
	Active      bool
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
}

type GroupMember struct {
	ID    string
	Name  string
	Email string
}

type GroupFiles struct {
	Logo          *File
	FinancingPlan *File
	GroupProject  *File
}
