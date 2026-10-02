package course_requests

// ApproveCourseRequestRequest is sent as JSON, or as multipart when it carries the evaluation
// document (file archivo_evaluacion).
type ApproveCourseRequestRequest struct {
	CourseType     string `json:"tipo_curso" form:"tipo_curso"`
	Comments       string `json:"observaciones" form:"observaciones"`
	Score          string `json:"calificacion" form:"calificacion"`
	Classification string `json:"clasificacion" form:"clasificacion"`
}

type RejectCourseRequestRequest struct {
	Comments string `json:"observaciones"`
}

type RedirectCourseRequestRequest struct {
	Faculty string `json:"facultad" validate:"required"`
	Reason  string `json:"motivo" validate:"required"`
}
