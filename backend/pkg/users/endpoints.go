package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Service interface {
	GetUser(ctx context.Context, userID string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int, error)
	UpdateUser(ctx context.Context, userID int, user entities.User) error
	DeleteUser(ctx context.Context, userID int) error
}

type Endpoints struct {
	GetUser    endpoint.Endpoint
	GetUsers   endpoint.Endpoint
	CreateUser endpoint.Endpoint
	UpdateUser endpoint.Endpoint
	DeleteUser endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		GetUser:    makeGetUser(svc, log),
		GetUsers:   makeGetUsers(svc, log),
		CreateUser: makeCreateUser(svc, log),
		UpdateUser: makeUpdateUser(svc, log),
		DeleteUser: makeDeleteUser(svc, log),
	}
}

func wrapEndpoint(e endpoint.Endpoint, middlewares []endpoint.Middleware) endpoint.Endpoint {
	for _, m := range middlewares {
		e = m(e)
	}
	return e
}

func makeGetUser(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetUserRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		user, err := svc.GetUser(ctx, req.UserID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := GetUserResponse{
			User: user,
		}
		return response, nil
	}
}

func makeGetUsers(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetUsersRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		users, pages, err := svc.GetUsers(ctx, req.PageScope)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := GetUsersResponse{
			Users: users,
			Pages: pages,
		}
		return response, nil
	}
}

func makeCreateUser(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(CreateUserRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		newUser := entities.User{
			Username: req.Username,
			Password: req.Password,
		}
		userID, err := svc.CreateUser(ctx, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := CreateUsersResponse{
			ID: fmt.Sprintf("%d", userID),
		}
		return response, nil
	}
}

func makeUpdateUser(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// req, ok := request
		return nil, nil
	}
}

func makeDeleteUser(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// req, ok := request
		return nil, nil
	}
}
