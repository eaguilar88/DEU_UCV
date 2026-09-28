package group_resource_requests

type RejectGroupResourceRequestRequest struct {
	Reason string `json:"razon" validate:"required"`
}
