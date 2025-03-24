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
	user := users.UserEntityToGetUserResponse(endorsement.User)
	reviewer := users.UserEntityToGetUserResponse(endorsement.Reviewer)
	return GetEndorsementResponse{
		ID:          endorsement.ID,
		User:        &user,
		Reviewer:    &reviewer,
		Status:      endorsement.Status,
		Type:        endorsement.Type,
		Name:        endorsement.Name,
		Description: endorsement.Description,
		Comments:    endorsement.Comments,
		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAtAt,
	}
}
