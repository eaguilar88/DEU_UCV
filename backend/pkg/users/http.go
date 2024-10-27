package users

import (
	"github.com/go-kit/kit/endpoint"
	kitHTTP "github.com/go-kit/kit/transport/http"
)

type swaggerVariables struct {
	Version     string
	Name        string
	GitCommitID string
}

// User Endpoints
func GetUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetUserRequestHTTP,
		encodeGetUserResponseHTTP,
		options...,
	)
}

func GetUsersHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetUsersRequestHTTP,
		encodeGetUsersResponseHTTP,
		options...,
	)
}

func CreateUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeCreateUserRequestHTTP,
		encodeCreateUserResponseHTTP,
		options...,
	)
}

func UpdateUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeUpdateUserRequestHTTP,
		encodeUpdateUserResponseHTTP,
		options...,
	)
}

func DeleteUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeDeleteUserRequestHTTP,
		encodeDeleteUserResponseHTTP,
		options...,
	)
}
