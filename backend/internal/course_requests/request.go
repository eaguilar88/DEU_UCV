package course_requests

type ApproveCourseRequestRequest struct {
	CourseType string `json:"tipo_curso"`
	Comments   string `json:"observaciones"`
}

type RejectCourseRequestRequest struct {
	Comments string `json:"observaciones"`
}

type RedirectCourseRequestRequest struct {
	Faculty string `json:"facultad" validate:"required"`
	Reason  string `json:"motivo" validate:"required"`
}
