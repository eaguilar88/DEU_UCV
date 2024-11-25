package auth

import (
	"context"
	"errors"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Service interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
}

type Endpoints struct {
	Login endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		Login: makeLogin(svc, log),
	}
}

func makeLogin(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(LoginRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		token, user, err := svc.Login(ctx, req.Username, req.Password)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := LoginResponse{
			Token: token,
			User: LoginUserResponse{
				ID:    user.ID,
				Name:  user.FirstName + " " + user.LastName,
				Roles: user.Roles,
			},
		}
		return response, nil
	}
}
