package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func encodeLoginResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(LoginResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}
