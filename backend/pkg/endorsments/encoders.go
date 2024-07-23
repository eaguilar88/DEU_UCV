package endorsments

import "github.com/eaguilar88/deu/pkg/entities"

func entitiesEndorsmentToGetEndorsmentResponse(endorsment entities.Endorsments) GetEndorsmentResponse {
	return GetEndorsmentResponse{
		ID:          endorsment.ID,
		User:        endorsment.User,
		Status:      "",
		Path:        endorsment.Path,
		CreatedAt:   endorsment.CreatedAt,
		UpdatedAtAt: endorsment.UpdatedAtAt,
	}
}
