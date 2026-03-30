package course_cycle_close_requests

type SubmitCloseRequestRequest struct {
	CourseCycleID int64  `json:"course_cycle_id" validate:"required"`
	Comments      string `json:"observaciones"`
}

type RejectCloseRequestRequest struct {
	Comments string `json:"observaciones"`
}
