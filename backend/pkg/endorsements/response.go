package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsementResponse struct {
	ID          int                        `json:"id,omitempty"`
	User        entities.User              `json:"user,omitempty"`
	Status      entities.EndorsementStatus `json:"status,omitempty"`
	Type        string                     `json:"type,omitempty"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	Comments    string                     `json:"comments,omitempty"`
	CreatedAt   string                     `json:"created_at,omitempty"`
	UpdatedAtAt string                     `json:"updated_at_at,omitempty"`
}

type GetEndorsementsResponse struct {
	Endorsements []entities.Endorsements `json:"requests"`
	Pages        entities.PageScope      `json:"pages"`
}

type CreateEndorsementResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateEndorsementResponse struct{}

type DeleteEndorsementResponse struct{}
