package endorsments

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsmentResponse struct {
	ID          int                       `json:"id,omitempty"`
	User        entities.User             `json:"user,omitempty"`
	Status      entities.EndorsmentStatus `json:"status,omitempty"`
	Path        string                    `json:"path,omitempty"`
	CreatedAt   string                    `json:"created_at,omitempty"`
	UpdatedAtAt string                    `json:"updated_at_at,omitempty"`
}

type GetEndorsmentsResponse struct {
	Endorsments []entities.Endorsments `json:"endorsment_requests,omitempty"`
	Pages       entities.PageScope     `json:"pages,omitempty"`
}

type CreateEndorsmentResponse struct {
	ID string `json:"id,omitempty"`
}

type UpdateEndorsmentResponse struct{}

type DeleteEndorsmentResponse struct{}
