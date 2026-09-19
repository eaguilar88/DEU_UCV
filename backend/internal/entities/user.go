package entities

import (
	"time"
)

type User struct {
	ID                 string
	CI                 string
	Email              string
	Roles              []string
	FirstName          string
	LastName           string
	DateOfBirth        string
	Age                int
	Gender             string
	EducationLevel     string
	ProviderCode       string
	Faculty            string
	Address            string
	Password           string
	CreatedAt          string
	ProfilePictureURL  string
	GroupID            string
	GroupName          string
	CourseProviderID   string
	CourseProviderName string
}

type UserRole struct {
	Name       string
	DomainType string
	Faculty    string
}

func (u *User) SetAge() {
	dob, err := time.Parse("2006-01-02", u.DateOfBirth)
	if err != nil {
		return
	}
	currentDate := time.Now()
	age := currentDate.Year() - dob.Year()
	if currentDate.Month() < dob.Month() ||
		(currentDate.Month() == dob.Month() && currentDate.Day() < dob.Day()) {
		age--
	}
	u.Age = age
}

// Role IDs and names mirror the rows seeded into deu.roles
// (postgresql/migrations/000020_seed_roles.up.sql) — that table is the source of truth, so
// these must stay in sync with it rather than the other way around.
const (
	RoleRoot = iota + 1
	RoleDeuAdmin
	RoleFacultyAdmin
	RoleCourseAdmin
	RoleCourseManager
	RoleGroupHelper
	RoleGroupAdmin
	RoleVisitante
)

var roleNames = map[int]string{
	RoleRoot:          "root",
	RoleDeuAdmin:      "deu_admin",
	RoleFacultyAdmin:  "faculty_admin",
	RoleCourseAdmin:   "course_admin",
	RoleCourseManager: "course_manager",
	RoleGroupHelper:   "group_helper",
	RoleGroupAdmin:    "group_admin",
	RoleVisitante:     "visitante",
}

var roleIDs = map[string]int{
	"root":           RoleRoot,
	"deu_admin":      RoleDeuAdmin,
	"faculty_admin":  RoleFacultyAdmin,
	"course_admin":   RoleCourseAdmin,
	"course_manager": RoleCourseManager,
	"group_helper":   RoleGroupHelper,
	"group_admin":    RoleGroupAdmin,
	"visitante":      RoleVisitante,
}

// RoleNameFromID returns the role name for a given ID.
func RoleNameFromID(id int) string {
	if name, ok := roleNames[id]; ok {
		return name
	}
	return ""
}

// RoleIDFromName returns the role ID for a given name.
func RoleIDFromName(name string) int {
	if id, ok := roleIDs[name]; ok {
		return id
	}
	return 0
}
