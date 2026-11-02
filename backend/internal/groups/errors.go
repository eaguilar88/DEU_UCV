package groups

import "errors"

var ErrGroupNotFound = errors.New("group not found")

var (
	// ErrNotGroupOwner is returned when someone other than the group's user submits its renewal.
	ErrNotGroupOwner = errors.New("only the group's user can renew it")
	// ErrRenewalNotOpen is returned when the group was never approved or its renewal date is not
	// within the renewal window yet.
	ErrRenewalNotOpen = errors.New("group renewal is not open yet")
	// ErrRenewalPending is returned when the group already has a renewal under review.
	ErrRenewalPending = errors.New("group already has a renewal under review")
)
