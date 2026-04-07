package entities

import (
	"time"
)

type User struct {
	ID                string
	CI                string
	Email             string
	Roles             []string
	FirstName         string
	LastName          string
	DateOfBirth       string
	Age               int
	Gender            string
	EducationLevel    string
	ProviderCode      string
	Faculty           string
	Address           string
	Password          string
	CreatedAt         string
	ProfilePictureURL string
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

type Role struct {
	ID   int
	Name string
}

const (
	RoleRoot = iota + 1
	RoleAdmin
	RoleFacultyAdmin
	RoleCoordinador
	RoleFacilitador
	RoleParticipante
	RoleExtension
	RoleVisitante
)

var roleNames = map[int]string{
	RoleRoot:         "root",
	RoleAdmin:        "admin",
	RoleFacultyAdmin: "administrador_facultad",
	RoleCoordinador:  "coordinador",
	RoleFacilitador:  "facilitador",
	RoleParticipante: "participante",
	RoleExtension:    "extensión",
	RoleVisitante:    "visitante",
}

var roleIDs = map[string]int{
	"root":                   RoleRoot,
	"admin":                  RoleAdmin,
	"administrador_facultad": RoleFacultyAdmin,
	"coordinador":            RoleCoordinador,
	"facilitador":            RoleFacilitador,
	"participante":           RoleParticipante,
	"extensión":              RoleExtension,
	"visitante":              RoleVisitante,
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
