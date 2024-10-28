package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeLoginRequestHTTP(ctx context.Context, r *http.Request) (interface{}, error) {
	req := LoginRequest{}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("error decoding request: %v", err)
	}
	return req, nil
}
