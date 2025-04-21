package endorsements

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/users"
)

func EndorsementEntitiesToGetEndorsementsResponse(
	endorsements []entities.Endorsement,
) []GetEndorsementResponse {
	responses := make([]GetEndorsementResponse, 0, len(endorsements))
	for _, endorsement := range endorsements {
		responses = append(responses, EntitiesEndorsementToGetEndorsementResponse(endorsement))
	}
	return responses
}

func EntitiesEndorsementToGetEndorsementResponse(
	endorsement entities.Endorsement,
) GetEndorsementResponse {
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
