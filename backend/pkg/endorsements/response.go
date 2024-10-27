package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsementResponse struct {
	ID          int                        `json:"id,omitempty"`
	User        entities.User              `json:"user,omitempty"`
	Status      entities.EndorsementStatus `json:"status,omitempty"`
	Path        string                     `json:"path,omitempty"`
	CreatedAt   string                     `json:"created_at,omitempty"`
	UpdatedAtAt string                     `json:"updated_at_at,omitempty"`
}

type GetEndorsementsResponse struct {
	Endorsements []entities.Endorsements `json:"endorsement_requests,omitempty"`
	Pages        entities.PageScope      `json:"pages,omitempty"`
}

type CreateEndorsementResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateEndorsementResponse struct{}

type DeleteEndorsementResponse struct{}
