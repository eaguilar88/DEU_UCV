package endorsements

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createEndorsementRequestToEntitiesEndorsement(req CreateEndorsementRequest) entities.Endorsements {
	return entities.Endorsements{
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.EndorsementStatus(req.Status),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Comments:    req.Comments,
		CreatedAt:   req.CreatedAt,
		UpdatedAtAt: req.UpdatedAtAt,
	}
}

func updateEndorsementRequestToEntitiesEndorsement(req UpdateEndorsementRequest, ID int) entities.Endorsements {

	return entities.Endorsements{
		ID: ID,
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.EndorsementStatus(req.Status),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Comments:    req.Comments,
	}
}
