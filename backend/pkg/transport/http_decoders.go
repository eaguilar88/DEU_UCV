package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/pkg/endorsments"
	"github.com/gorilla/mux"
)

// Endorsment Decoders
func decodeGetEndorsmentRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsmentID, exists := mux.Vars(r)[ParamEndorsmentID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsmentID)
	}

	return endorsments.GetEndorsmentRequest{
		ID: EndorsmentID,
	}, nil
}

func decodeGetEndorsmentsRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	queryScope, err := NewQueryScopeFromURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("error decoding query query string: %v", err)
	}

	return endorsments.GetEndorsmentsRequest{
		PageScope: queryScope,
	}, nil
}

func decodeCreateEndorsmentRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	req := endorsments.CreateEndorsmentRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	return req, nil
}

func decodeUpdateEndorsmentRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsmentID, exists := mux.Vars(r)[ParamEndorsmentID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsmentID)
	}

	req := endorsments.UpdateEndorsmentRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	req.ID = EndorsmentID
	return req, nil
}

func decodeDeleteEndorsmentRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	EndorsmentID, exists := mux.Vars(r)[ParamEndorsmentID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamEndorsmentID)
	}

	return endorsments.DeleteEndorsmentRequest{
		ID: EndorsmentID,
	}, nil
}
