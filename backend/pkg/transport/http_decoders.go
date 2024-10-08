package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/pkg/users"
	"github.com/gorilla/mux"
)

func decodeGetUserRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	userID, exists := mux.Vars(r)[ParamUserID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamUserID)
	}

	return users.GetUserRequest{
		ID: userID,
	}, nil
}

func decodeGetUsersRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	queryScope, err := NewQueryScopeFromURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("error decoding query query string: %v", err)
	}

	return users.GetUsersRequest{
		PageScope: queryScope,
	}, nil
}

func decodeCreateUserRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	req := users.CreateUserRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	return req, nil
}

func decodeUpdateUserRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	userID, exists := mux.Vars(r)[ParamUserID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamUserID)
	}

	req := users.UpdateUserRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	req.ID = userID
	return req, nil
}

func decodeDeleteUserRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	userID, exists := mux.Vars(r)[ParamUserID]
	if !exists {
		return nil, fmt.Errorf("missing required param: %s", ParamUserID)
	}

	return users.DeleteUserRequest{
		ID: userID,
	}, nil
}
