package entities

type Course struct {
	ID         int         `json:"id,omitempty"`
	Owner      User        `json:"owner,omitempty"`
	Endorsment Endorsments `json:"endorsment,omitempty"`
	SignedBy   string      `json:"signed_by,omitempty"`
	Objectives string      `json:"objectives,omitempty"`
	Cost       float64     `json:"cost,omitempty"`
	Content    string      `json:"content,omitempty"`
	Location   string      `json:"location,omitempty"`
	CreatedAt  string      `json:"created_at,omitempty"`
	UpdatedAt  string      `json:"updated_at,omitempty"`
}
