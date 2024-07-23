package endorsments

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
	GetEndorsment(ctx context.Context, endorsmentID string) (entities.Endorsments, error)
	GetEndorsments(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsments, entities.PageScope, error)
	CreateEndorsment(ctx context.Context, Endorsment entities.Endorsments) (int64, error)
	UpdateEndorsment(ctx context.Context, endorsmentID int, Endorsment entities.Endorsments) error
	DeleteEndorsment(ctx context.Context, endorsmentID int) error
}

type Endpoints struct {
	GetEndorsment    endpoint.Endpoint
	GetEndorsments   endpoint.Endpoint
	CreateEndorsment endpoint.Endpoint
	UpdateEndorsment endpoint.Endpoint
	DeleteEndorsment endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		GetEndorsment:    makeGetEndorsment(svc, log),
		GetEndorsments:   makeGetEndorsments(svc, log),
		CreateEndorsment: makeCreateEndorsment(svc, log),
		UpdateEndorsment: makeUpdateEndorsment(svc, log),
		DeleteEndorsment: makeDeleteEndorsment(svc, log),
	}
}

func wrapEndpoint(e endpoint.Endpoint, middlewares []endpoint.Middleware) endpoint.Endpoint {
	for _, m := range middlewares {
		e = m(e)
	}
	return e
}

func makeGetEndorsment(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetEndorsmentRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		endorsment, err := svc.GetEndorsment(ctx, req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return entitiesEndorsmentToGetEndorsmentResponse(endorsment), nil
	}
}

func makeGetEndorsments(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetEndorsmentsRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		users, pages, err := svc.GetEndorsments(ctx, req.PageScope)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := GetEndorsmentsResponse{
			Endorsments: users,
			Pages:       pages,
		}
		return response, nil
	}
}

func makeCreateEndorsment(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(CreateEndorsmentRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		newUser := createEndorsmentRequestToEntitiesEndorsment(req)
		userID, err := svc.CreateEndorsment(ctx, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := CreateEndorsmentResponse{
			ID: fmt.Sprintf("%d", userID),
		}
		return response, nil
	}
}

func makeUpdateEndorsment(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateEndorsmentRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		newUser := updateEndorsmentRequestToEntitiesEndorsment(req, intID)
		err = svc.UpdateEndorsment(ctx, intID, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return UpdateEndorsmentResponse{}, nil
	}
}

func makeDeleteEndorsment(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(DeleteEndorsmentRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		err = svc.DeleteEndorsment(ctx, intID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return DeleteEndorsmentResponse{}, nil
	}
}
