package endorsments

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createEndorsmentRequestToEntitiesEndorsment(req CreateEndorsmentRequest) entities.Endorsments {
	return entities.Endorsments{}
}

func updateEndorsmentRequestToEntitiesEndorsment(req UpdateEndorsmentRequest, ID int) entities.Endorsments {

	return entities.Endorsments{
		ID: ID,
	}
}
