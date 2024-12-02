package auth

import (
	"github.com/go-kit/kit/endpoint"
	kitHTTP "github.com/go-kit/kit/transport/http"
)

func LoginHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeLoginRequestHTTP,
		encodeLoginResponseHTTP,
		options...,
	)
}
