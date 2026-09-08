package entities

type GroupType string

const (
	// Group types
	CulturalGroupType          = "cultural"
	SportsGroupType            = "sports"
	AcademicGroupType          = "academic"
	MultidisciplinaryGroupType = "multidisciplinary"

	// File types for groups
	GroupFileTypeLogo    = "logo"
	GroupFileTypeProject = "proyecto_grupo"
	GroupMemberFileTypeDocument = "documento_miembro"
)

// ValidGroupTypes contains all valid group type values
//var ValidGroupTypes = map[GroupType]bool{
//	CulturalGroupType:          true,
//	SportsGroupType:            true,
//	AcademicGroupType:          true,
//	MultidisciplinaryGroupType: true,
//}

// IsValid checks if the GroupType value is valid
func (t GroupType) IsValid() bool {
	return string(t) != ""
}

// GroupFilter carries the optional filters accepted by GetGroups.
type GroupFilter struct {
	Faculty Faculty
	Type    GroupType
	Active  *bool
	Search  string
	Deleted bool
}

type ExtensionGroup struct {
	ID                  string
	Name                string
	Owner               *User
	LeadName            string
	Description         string
	Foundation          string
	IsMultidisciplinary bool
	Objective           string
	Location            string
	Type                []GroupType
	Faculty             []Faculty
	Logo                *File
	Project             *File
	Files               []*File
	Members             []GroupMember
	Active              bool
	Email               string
	Phone               string
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           string
}

type GroupMember struct {
	ID           string
	Name         string
	CI           int
	Phone        string
	Email        string
	Coordination string
	Year         string
	Faculty      Faculty
	School       string
	IsLeader     bool
	IsActive     bool
	Document     *File
}

type GroupFiles struct {
	Logo *File
	// FinancingPlan *File
	GroupProject *File
}
