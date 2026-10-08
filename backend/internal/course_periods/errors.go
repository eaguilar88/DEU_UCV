package course_periods

import "errors"

var (
	ErrCoursePeriodNotFound    = errors.New("course period not found")
	ErrAnnouncementNotFound    = errors.New("announcement not found")
	ErrCoursePeriodAlreadyOpen = errors.New("a course period is already open for this course")
	ErrCourseNotCovered        = errors.New("course is not covered by a legal contract")
	ErrInvalidCapacity         = errors.New("capacity must be greater than zero")
	ErrCohortNameRequired      = errors.New("nombre_cohorte is required")
)
