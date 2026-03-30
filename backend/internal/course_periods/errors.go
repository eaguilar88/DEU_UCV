package course_periods

import "errors"

var (
	ErrCoursePeriodNotFound = errors.New("course period not found")
	ErrAnnouncementNotFound = errors.New("announcement not found")
)
