package endorsements

import (
	"github.com/eaguilar88/deu/internal/entities"
)

func createEndorsementRequestToEntitiesEndorsement(
	req CreateEndorsementRequest,
) entities.Endorsement {
	return entities.Endorsement{
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

func updateEndorsementRequestToEntitiesEndorsement(
	req UpdateEndorsementRequest,
) entities.Endorsement {
	return entities.Endorsement{
		ID: req.ID,
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
