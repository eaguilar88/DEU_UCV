package group_requests

type RejectGroupRequestRequest struct {
	Reason string `json:"razon" validate:"required"`
}
