package entities

import (
	"time"
)

type User struct {
	ID             string
	CI             string
	Username       string
	Roles          []string
	FirstName      string
	LastName       string
	DateOfBirth    string
	Age            int
	Gender         string
	EducationLevel string
	ProviderCode   string
	Address        string
	Password       string
	CreatedAt      string
}

func (u *User) SetAge() {
	dob, err := time.Parse(time.RFC3339, u.DateOfBirth)
	if err != nil {
		return
	}
	currentDate := time.Now()
	age := currentDate.Year() - dob.Year()
	if currentDate.Month() < dob.Month() || (currentDate.Month() == dob.Month() && currentDate.Day() < dob.Day()) {
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
	RoleCoordinador
	RoleFacilitador
	RoleParticipante
	RoleExtension
)

var roleNames = map[int]string{
	RoleRoot:         "root",
	RoleAdmin:        "admin",
	RoleCoordinador:  "coordinador",
	RoleFacilitador:  "facilitador",
	RoleParticipante: "participante",
	RoleExtension:    "extensión",
}

var roleIDs = map[string]int{
	"root":         RoleRoot,
	"admin":        RoleAdmin,
	"coordinador":  RoleCoordinador,
	"facilitador":  RoleFacilitador,
	"participante": RoleParticipante,
	"extensión":    RoleExtension,
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
