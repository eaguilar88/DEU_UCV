package entities

import "time"

type User struct {
	ID             int    `json:"id"`
	CI             string `json:"ci"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	EducationLevel string `json:"education_level"`
	Address        string `json:"address"`
	Password       string `json:"password"`
	CreatedAt      string `json:"created_at"`
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
