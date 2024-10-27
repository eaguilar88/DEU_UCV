package endorsements

import "github.com/eaguilar88/deu/pkg/entities"

func entitiesEndorsementToGetEndorsementResponse(endorsement entities.Endorsements) GetEndorsementResponse {
	return GetEndorsementResponse{
		ID:          endorsement.ID,
		User:        endorsement.User,
		Status:      "",
		Path:        endorsement.Path,
		CreatedAt:   endorsement.CreatedAt,
		UpdatedAtAt: endorsement.UpdatedAtAt,
	}
}
