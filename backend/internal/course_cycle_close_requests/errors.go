package course_cycle_close_requests

import "errors"

var (
	ErrCycleCloseRequestNotFound  = errors.New("course cycle close request not found")
	ErrCloseRequestAlreadyPending = errors.New("a close request is already pending review for this course cycle")
)
