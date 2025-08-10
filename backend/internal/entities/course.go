package entities

const (
	CourseFileTypeLogoEntryProfile       = "entry_profile"
	CourseFileTypeLogoExitProfile        = "exit_profile"
	CourseFileTypeLogoFacilitatorProfile = "facilitator_profile"
	CourseFileTypeCostStructure          = "cost_structure"
	CourseFileTypeYearlyPlan             = "yearly_plan"
)

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
	Files         CourseFiles
	CreatedAt     string
	UpdatedAt     string
}

type CourseFiles struct {
	EntryProfile       *File
	ExitProfile        *File
	FacilitatorProfile *File
}
