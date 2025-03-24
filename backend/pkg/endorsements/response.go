package endorsements

import (
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
)

type GetEndorsementResponse struct {
	ID          int                        `json:"id,omitempty"`
	User        *users.GetUserResponse     `json:"user,omitempty"`
	Reviewer    *users.GetUserResponse     `json:"reviewer,omitempty"`
	Status      entities.EndorsementStatus `json:"status,omitempty"`
	Type        string                     `json:"type,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	Comments    string                     `json:"comments,omitempty"`
	ReviewedAt  string                     `json:"reviewed_at,omitempty"`
	CreatedAt   string                     `json:"created_at,omitempty"`
	UpdatedAtAt string                     `json:"updated_at_at,omitempty"`
}

type GetEndorsementsResponse struct {
	Endorsements []GetEndorsementResponse `json:"requests"`
	Pages        entities.PageScope       `json:"pages"`
}

type CreateEndorsementResponse struct {
	ID string `json:"id,omitempty"`
}
