package users

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Service interface {
	GetUser(ctx context.Context, userID string) (entities.User, error)
	GetUsers(ctx context.Context, pageScope entities.PageScope) ([]entities.User, entities.PageScope, error)
	CreateUser(ctx context.Context, user entities.User) (int64, error)
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
		user, err := svc.GetUser(ctx, req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return entitiesUserToGetUserResponse(user), nil
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
		newUser := createUserRequestToEntitiesUser(req)
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
		req, ok := request.(UpdateUserRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		newUser := updateUserRequestToEntitiesUser(req, intID)
		err = svc.UpdateUser(ctx, intID, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return UpdateUserResponse{}, nil
	}
}

func makeDeleteUser(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(DeleteUserRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		err = svc.DeleteUser(ctx, intID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return DeleteUserResponse{}, nil
	}
}
