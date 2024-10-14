package auth

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, password string) (string, error)
}

type Endpoints struct {
	Login    endpoint.Endpoint
	Register endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		Login:    makeLogin(svc, log),
		Register: makeRegister(svc, log),
	}
}

func makeLogin(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(LoginRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		token, err := svc.Login(ctx, req.Username, req.Password)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := LoginResponse{
			Token: token,
		}
		return response, nil
	}
}
func makeRegister(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(RegisterRequest)
		if !ok {
			return nil, errors.New("could not decode")
		}
		token, err := svc.Register(ctx, req.Username, req.Password)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := RegisterResponse{
			ID: token,
		}
		return response, nil
	}
}
