package entities

import "errors"

type CourseType string

const (
	CourseType_Undefined         CourseType = "unassigned"
	CourseType_SkillDevelopment  CourseType = "skill_development"
	CourseType_LifeSkills        CourseType = "life_skills"
	CourseType_TechnicalTraining CourseType = "technical_training"

	//Files
	CourseFileTypeCover = "portada"
)

// CourseManagementStatus reflects course-level lifecycle state that isn't captured by CourseType.
// The empty value means no special status (approved but never yet opened).
type CourseManagementStatus string

const (
	CourseManagementStatusClosureRequested CourseManagementStatus = "solicitud-cierre"
	CourseManagementStatusOpen             CourseManagementStatus = "abierto"
	CourseManagementStatusClosed           CourseManagementStatus = "cerrado"
)

var (
	ErrInvalidCourseType = errors.New("invalid course type")
)

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
	Location          string
	Objectives        string
	Owner             User
	Periods           []CoursePeriod
	Type              CourseType
	HasDocumentation  bool
	ManagementStatus  CourseManagementStatus
	CreatedAt         string
	UpdatedAt         string
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
