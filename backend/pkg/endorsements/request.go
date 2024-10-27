package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

type GetEndorsementsRequest struct {
	PageScope entities.PageScope
}

type GetEndorsementRequest struct {
	ID string
}

type CreateEndorsementRequest struct {
	ID     int                        `json:"id,omitempty"`
	User   entities.User              `json:"user,omitempty"`
	Status entities.EndorsementStatus `json:"status,omitempty"`
	Path   string                     `json:"path,omitempty"`
}

type DeleteEndorsementRequest struct {
	ID string
}

type UpdateEndorsementRequest struct {
	ID     string
	User   entities.User              `json:"user,omitempty"`
	Status entities.EndorsementStatus `json:"status,omitempty"`
	Path   string                     `json:"path,omitempty"`
}
