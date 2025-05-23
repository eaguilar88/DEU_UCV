package entities

type Course struct {
	ID            string
	Content       string
	Cost          float64
	Description   string
	CourseRequest CourseRequest
	Location      string
	Name          string
	Objectives    string
	Owner         User
	Periods       []CoursePeriod
	CreatedAt     string
	UpdatedAt     string
}
