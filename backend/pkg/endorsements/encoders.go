package endorsements

import (
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
)

func endorsementEntitiesToGetEndorsementsResponse(endorsements []entities.Endorsements) []GetEndorsementResponse {
	var responses = make([]GetEndorsementResponse, 0, len(endorsements))
	for _, endorsement := range endorsements {
		responses = append(responses, entitiesEndorsementToGetEndorsementResponse(endorsement))
	}
	return responses
}

func entitiesEndorsementToGetEndorsementResponse(endorsement entities.Endorsements) GetEndorsementResponse {
	return GetEndorsementResponse{
		ID:          endorsement.ID,
		User:        users.UserEntityToGetUserResponse(endorsement.User),
		Reviewer:    users.UserEntityToGetUserResponse(endorsement.Reviewer),
		Status:      endorsement.Status,
		Type:        endorsement.Type,
		Name:        endorsement.Name,
		Description: endorsement.Description,
		Comments:    endorsement.Comments,
		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAtAt,
	}
}
