package endorsements

import (
	"github.com/eaguilar88/deu/pkg/entities"
)

func createEndorsementRequestToEntitiesEndorsement(req CreateEndorsementRequest) entities.Endorsements {
	return entities.Endorsements{}
}

func updateEndorsementRequestToEntitiesEndorsement(req UpdateEndorsementRequest, ID int) entities.Endorsements {

	return entities.Endorsements{
		ID: ID,
	}
}
