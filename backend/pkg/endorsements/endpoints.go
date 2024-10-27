package endorsements

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
	GetEndorsement(ctx context.Context, endorsementID string) (entities.Endorsements, error)
	GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsements, entities.PageScope, error)
	CreateEndorsement(ctx context.Context, endorsement entities.Endorsements) (int64, error)
	UpdateEndorsement(ctx context.Context, endorsementID int, Endorsement entities.Endorsements) error
	DeleteEndorsement(ctx context.Context, endorsementID int) error
}

type Endpoints struct {
	GetEndorsement    endpoint.Endpoint
	GetEndorsements   endpoint.Endpoint
	CreateEndorsement endpoint.Endpoint
	UpdateEndorsement endpoint.Endpoint
	DeleteEndorsement endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		GetEndorsement:    makeGetEndorsement(svc, log),
		GetEndorsements:   makeGetEndorsements(svc, log),
		CreateEndorsement: makeCreateEndorsement(svc, log),
		UpdateEndorsement: makeUpdateEndorsement(svc, log),
		DeleteEndorsement: makeDeleteEndorsement(svc, log),
	}
}

func wrapEndpoint(e endpoint.Endpoint, middlewares []endpoint.Middleware) endpoint.Endpoint {
	for _, m := range middlewares {
		e = m(e)
	}
	return e
}

func makeGetEndorsement(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetEndorsementRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		endorsement, err := svc.GetEndorsement(ctx, req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return entitiesEndorsementToGetEndorsementResponse(endorsement), nil
	}
}

func makeGetEndorsements(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetEndorsementsRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		users, pages, err := svc.GetEndorsements(ctx, req.PageScope)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := GetEndorsementsResponse{
			Endorsements: users,
			Pages:        pages,
		}
		return response, nil
	}
}

func makeCreateEndorsement(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(CreateEndorsementRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		newUser := createEndorsementRequestToEntitiesEndorsement(req)
		userID, err := svc.CreateEndorsement(ctx, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := CreateEndorsementResponse{
			ID: fmt.Sprintf("%d", userID),
		}
		return response, nil
	}
}

func makeUpdateEndorsement(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateEndorsementRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		newUser := updateEndorsementRequestToEntitiesEndorsement(req, intID)
		err = svc.UpdateEndorsement(ctx, intID, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return UpdateEndorsementResponse{}, nil
	}
}

func makeDeleteEndorsement(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(DeleteEndorsementRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		err = svc.DeleteEndorsement(ctx, intID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return DeleteEndorsementResponse{}, nil
	}
}
