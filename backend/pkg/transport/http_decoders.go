package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/pkg/endorsements"
	"github.com/gorilla/mux"
)

// Endorsement Decoders
func decodeGetEndorsementRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsementID, exists := mux.Vars(r)[ParamEndorsementID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsementID)
	}

	return endorsements.GetEndorsementRequest{
		ID: EndorsementID,
	}, nil
}

func decodeGetEndorsementsRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	queryScope, err := NewQueryScopeFromURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("error decoding query query string: %v", err)
	}

	return endorsements.GetEndorsementsRequest{
		PageScope: queryScope,
	}, nil
}

func decodeCreateEndorsementRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	req := endorsements.CreateEndorsementRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	return req, nil
}

func decodeUpdateEndorsementRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsementID, exists := mux.Vars(r)[ParamEndorsementID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsementID)
	}

	req := endorsements.UpdateEndorsementRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	req.ID = EndorsementID
	return req, nil
}

func decodeDeleteEndorsementRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsementID, exists := mux.Vars(r)[ParamEndorsementID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsementID)
	}

	return endorsements.DeleteEndorsementRequest{
		ID: EndorsementID,
	}, nil
}
