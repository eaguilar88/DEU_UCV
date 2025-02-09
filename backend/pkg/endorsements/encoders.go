package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

func entitiesEndorsementToGetEndorsementResponse(endorsement entities.Endorsements) GetEndorsementResponse {
	return GetEndorsementResponse{
		ID:          endorsement.ID,
		User:        endorsement.User,
		Status:      endorsement.Status,
		Type:        endorsement.Type,
		Name:        endorsement.Name,
		Description: endorsement.Description,
		Comments:    endorsement.Comments,
		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAtAt,
	}
}
