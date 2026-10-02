package entities

import "errors"

type CourseType string

const (
	CourseType_Undefined         CourseType = "unassigned"
	CourseType_SkillDevelopment  CourseType = "skill_development"
	CourseType_LifeSkills        CourseType = "life_skills"
	CourseType_TechnicalTraining CourseType = "technical_training"

	//Files
	CourseFileTypeCover         = "portada"
	CourseFileTypeFacilitatorCV = "cv_facilitador"
)

// CourseManagementStatus reflects course-level lifecycle state that isn't captured by CourseType.
// The empty value means no special status (approved but never yet opened).
type CourseManagementStatus string

const (
	CourseManagementStatusClosureRequested CourseManagementStatus = "solicitud-cierre"
	CourseManagementStatusOpen             CourseManagementStatus = "abierto"
	CourseManagementStatusClosed           CourseManagementStatus = "cerrado"
)

func (s CourseManagementStatus) IsValid() bool {
	switch s {
	case CourseManagementStatusClosureRequested, CourseManagementStatusOpen, CourseManagementStatusClosed:
		return true
	}
	return false
}

// CourseFilter narrows GET /courses. VisibleStatuses is set by the service, never by the
// client: when non-empty, only courses in one of those management statuses are returned.
type CourseFilter struct {
	OwnerUserID      string
	ProviderCode     string
	ManagementStatus CourseManagementStatus
	VisibleStatuses  []CourseManagementStatus
}

var (
	ErrInvalidCourseType = errors.New("invalid course type")
)

// CourseProviderSummary is what a course page shows publicly about its provider.
type CourseProviderSummary struct {
	// UserID is the provider's user account, i.e. the course owner.
	UserID  string
	Name    string
	LogoURL string
}

type Course struct {
	ID                string
	Name              string
	Description       string
	Cover             *File
	Content           string
	Cost              string
	Rationale         string
	InstructorProfile string
	Profiles          string
	Requirements      string
	Evaluation        string
	Schedule          string
	CourseRequest     CourseRequest
	Duration          string
	Faculty           Faculty
	OriginFaculty     Faculty
	Location          string
	Objectives        string
	Owner             User
	Periods           []CoursePeriod
	Type              CourseType
	// IsActive is true once the course request is approved.
	IsActive         bool
	HasDocumentation bool
	ManagementStatus CourseManagementStatus
	// Competencies lists the course modules with their contents and competencies.
	Competencies  string
	Bibliography  string
	FacilitatorCV *File
	// Provider is the public profile of the course owner, loaded only on the course detail.
	Provider  *CourseProviderSummary
	CreatedAt string
	UpdatedAt string
}

func (ct CourseType) IsValid() bool {
	switch ct {
	case CourseType_Undefined, CourseType_SkillDevelopment, CourseType_LifeSkills, CourseType_TechnicalTraining:
		return true
	default:
		return false
	}
}

func (ct CourseType) String() string {
	return string(ct)
}

func FromStringCourseType(s string) CourseType {
	ct := CourseType(s)
	if !ct.IsValid() {
		return CourseType_Undefined
	}
	return ct
}
