package entities

const (
	GroupFileTypeLogo = "logo"
)

type ExtensionGroup struct {
	ID            string
	Name          string
	Description   string
	Owner         User
	CourseRequest CourseRequest
	Members       []GroupMember
	Objective     string
	Location      string
	Active        bool
	CreatedAt     string
	UpdatedAt     string
	DeletedAt     string
}

type GroupMember struct {
	ID    string
	Name  string
	Email string
}
