package entities

type GroupRequest struct {
	ID        string
	GroupID   string
	GroupName string
	// RenewalID is empty for the requests created with the group, and the ID of the yearly
	// renewal the request reviews otherwise.
	RenewalID  string
	Faculty    Faculty
	Status     RequestStatus
	Reviewer   *User
	Comments   string
	ReviewedAt string
	CreatedAt  string
	UpdatedAt  string
	Approvals  []GroupRequest
	// Renewal is the proposed group data, loaded only for a renewal request's detail.
	Renewal *GroupRenewal
}

// GroupRenewal is a group's yearly renewal: the whole group data (Group, including members, logo,
// project and member documents) submitted again by the group's user. It replaces the group's data
// once every renewal request is approved.
type GroupRenewal struct {
	ID          string
	GroupID     string
	Group       ExtensionGroup
	Status      RequestStatus
	SubmittedBy string
	CreatedAt   string
}

// GroupRenewalMemberIndexKey is the file metadata key holding the position, in the renewal's
// member list, of the member a proposed document belongs to.
const GroupRenewalMemberIndexKey = "member_index"

type FacultyPendingCount struct {
	Faculty Faculty `json:"facultad"`
	Count   int     `json:"pendientes"`
}
