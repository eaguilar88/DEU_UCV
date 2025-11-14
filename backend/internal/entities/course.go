package entities

import "errors"

type CourseType string

const (
	CourseType_Undefined         CourseType = "unassigned"
	CourseType_SkillDevelopment  CourseType = "skill_development"
	CourseType_LifeSkills        CourseType = "life_skills"
	CourseType_TechnicalTraining CourseType = "technical_training"
)

var (
	ErrInvalidCourseType = errors.New("invalid course type")
)

type Course struct {
	ID            string
	Content       string
	Cost          float64
	Description   string
	CourseRequest CourseRequest
	Duration      int
	Faculty       Faculty
	Location      string
	Name          string
	Objectives    string
	Owner         User
	Periods       []CoursePeriod
	Type          CourseType
	CreatedAt     string
	UpdatedAt     string
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
