package endorsements

import (
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
)

func EndorsementEntitiesToGetEndorsementsResponse(endorsements []entities.Endorsements) []GetEndorsementResponse {
	var responses = make([]GetEndorsementResponse, 0, len(endorsements))
	for _, endorsement := range endorsements {
		responses = append(responses, EntitiesEndorsementToGetEndorsementResponse(endorsement))
	}
	return responses
}

func EntitiesEndorsementToGetEndorsementResponse(endorsement entities.Endorsements) GetEndorsementResponse {
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
