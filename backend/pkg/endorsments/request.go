package endorsments

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsmentsRequest struct {
	PageScope entities.PageScope
}

type GetEndorsmentRequest struct {
	ID string
}

type CreateEndorsmentRequest struct {
	ID     int                       `json:"id,omitempty"`
	User   entities.User             `json:"user,omitempty"`
	Status entities.EndorsmentStatus `json:"status,omitempty"`
	Path   string                    `json:"path,omitempty"`
}

type DeleteEndorsmentRequest struct {
	ID string
}

type UpdateEndorsmentRequest struct {
	ID     string
	User   entities.User             `json:"user,omitempty"`
	Status entities.EndorsmentStatus `json:"status,omitempty"`
	Path   string                    `json:"path,omitempty"`
}
