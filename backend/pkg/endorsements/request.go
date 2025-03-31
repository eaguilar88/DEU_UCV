package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsementsRequest struct {
	PageScope entities.PageScope
}

type GetEndorsementRequest struct {
	ID string
}

type CreateEndorsementRequest struct {
	UserID      string `json:"user_id,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Comments    string `json:"comments,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAtAt string `json:"updated_at_at,omitempty"`
}

type DeleteEndorsementRequest struct {
	ID string `path:"id"`
}

type UpdateEndorsementRequest struct {
	ID          string `path:"id"`
	UserID      string `json:"user_id,omitempty"`
	Type        string `json:"type,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Comments    string `json:"comments,omitempty"`
}
