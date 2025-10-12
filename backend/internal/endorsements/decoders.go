package endorsements

import (
	"github.com/eaguilar88/deu/internal/entities"
)

func createEndorsementRequestToEntitiesEndorsement(
	req CreateEndorsementRequest,
) entities.CourseRequest {
	return entities.CourseRequest{
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.RequestStatus(req.Status),
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
) entities.CourseRequest {
	return entities.CourseRequest{
		ID: req.ID,
		User: entities.User{
			ID: req.UserID,
		},
		Status:      entities.RequestStatus(req.Status),
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Comments:    req.Comments,
	}
}
