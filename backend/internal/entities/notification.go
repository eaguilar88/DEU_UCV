package entities

import "time"

// Kinds of reminders sent by the notifier service, as stored in deu.notifications_log.
const (
	NotificationKindCoordinatorPendingRequests = "coordinator_pending_requests"
	NotificationKindGroupRenewal               = "group_renewal"
)

// NotificationLog records a reminder of Kind sent to Recipient about Target (a faculty code or a
// group ID).
type NotificationLog struct {
	Kind      string
	Target    string
	Recipient string
}

// GroupRenewalDue is a group whose yearly registration is due at RenewalDueAt.
type GroupRenewalDue struct {
	GroupID      string
	GroupName    string
	RenewalDueAt time.Time
}
