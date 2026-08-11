package activities

import "errors"

var (
	ErrActivityNotFound        = errors.New("activity not found")
	ErrMaxFeaturedLimitReached = errors.New("maximum featured activities limit reached for this group")
)
