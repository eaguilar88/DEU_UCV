package group_requests

import "github.com/eaguilar88/deu/internal/entities"

type GetGroupRequestResponse struct {
	ID        string `json:"id"`
	GroupID   string `json:"group_id"`
	Comments  string `json:"comments"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type GetGroupRequestsResponse struct {
	Requests []GetGroupRequestResponse `json:"requests"`
	Pages    entities.PageScope        `json:"pages"`
}
