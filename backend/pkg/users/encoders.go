package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func encodeGetUserResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	resp, ok := untypedResp.(GetUserResponse)
	if !ok {
		return errors.New("dang bang")
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errors.New("fail again")
	}
	w.Header().Set("Content-Type", "application/jjson")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
	return nil
}

func encodeGetUsersResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	return nil
}

func encodeCreateUserResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	return nil
}

func encodeUpdateUserResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	return nil
}

func encodeDeleteUserResponseHTTP(_ context.Context, w http.ResponseWriter, untypedResp interface{}) error {
	return nil
}
