package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/kit/endpoint"
	kitHTTP "github.com/go-kit/kit/transport/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func LoginHandler(ctx context.Context, service *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Login Completed")
	}
}

func NewQueryScopeFromURL(url *url.URL) (entities.PageScope, error) {
	scope := entities.PageScope{}
	vars := url.Query()
	if err := scope.GetPageFromVars(vars.Get(PageParam)); err != nil {
		return scope, fmt.Errorf("error getting page from query string: %v", err)
	}

	if err := scope.GetPerPageFromVars(vars.Get(PerPageParam)); err != nil {
		return scope, fmt.Errorf("error getting per page from query string: %v", err)
	}

	return scope, nil
}

type swaggerVariables struct {
	Version     string
	Name        string
	GitCommitID string
}

// // User Endpoints
// func GetUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
// 	return kitHTTP.NewServer(
// 		ep,
// 		decodeGetUserRequestHTTP,
// 		encodeGetUserResponseHTTP,
// 		options...,
// 	)
// }

// func GetUsersHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
// 	return kitHTTP.NewServer(
// 		ep,
// 		decodeGetUsersRequestHTTP,
// 		encodeGetUsersResponseHTTP,
// 		options...,
// 	)
// }

// func CreateUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
// 	return kitHTTP.NewServer(
// 		ep,
// 		decodeCreateUserRequestHTTP,
// 		encodeCreateUserResponseHTTP,
// 		options...,
// 	)
// }

// func UpdateUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
// 	return kitHTTP.NewServer(
// 		ep,
// 		decodeUpdateUserRequestHTTP,
// 		encodeUpdateUserResponseHTTP,
// 		options...,
// 	)
// }

// func DeleteUserHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
// 	return kitHTTP.NewServer(
// 		ep,
// 		decodeDeleteUserRequestHTTP,
// 		encodeDeleteUserResponseHTTP,
// 		options...,
// 	)
// }

// Endorsement Endpoints
func GetEndorsementHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetEndorsementRequestHTTP,
		encodeGetEndorsementResponseHTTP,
		options...,
	)
}

func GetEndorsementsHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeGetEndorsementsRequestHTTP,
		encodeGetEndorsementsResponseHTTP,
		options...,
	)
}

func CreateEndorsementHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeCreateEndorsementRequestHTTP,
		encodeCreateEndorsementResponseHTTP,
		options...,
	)
}

func UpdateEndorsementHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeUpdateEndorsementRequestHTTP,
		encodeUpdateEndorsementResponseHTTP,
		options...,
	)
}

func DeleteEndorsementHandleHTTP(ep endpoint.Endpoint, options []kitHTTP.ServerOption) *kitHTTP.Server {
	return kitHTTP.NewServer(
		ep,
		decodeDeleteEndorsementRequestHTTP,
		encodeDeleteEndorsementResponseHTTP,
		options...,
	)
}
