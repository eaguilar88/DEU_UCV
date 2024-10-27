package courses

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
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	CreateCourse(ctx context.Context, user entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID int, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID int) error
}

type Endpoints struct {
	GetCourse    endpoint.Endpoint
	GetCourses   endpoint.Endpoint
	CreateCourse endpoint.Endpoint
	UpdateCourse endpoint.Endpoint
	DeleteCourse endpoint.Endpoint
}

func MakeEndpoints(svc Service, log log.Logger, middlewares ...endpoint.Middleware) Endpoints {
	return Endpoints{
		GetCourse:    makeGetCourse(svc, log),
		GetCourses:   makeGetCourses(svc, log),
		CreateCourse: makeCreateCourse(svc, log),
		UpdateCourse: makeUpdateCourse(svc, log),
		DeleteCourse: makeDeleteCourse(svc, log),
	}
}

func wrapEndpoint(e endpoint.Endpoint, middlewares []endpoint.Middleware) endpoint.Endpoint {
	for _, m := range middlewares {
		e = m(e)
	}
	return e
}

func makeGetCourse(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetCourseRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		user, err := svc.GetCourse(ctx, req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return entitiesCourseToGetCourseResponse(user), nil
	}
}

func makeGetCourses(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(GetCoursesRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		users, pages, err := svc.GetCourses(ctx, req.PageScope)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := GetCoursesResponse{
			Courses: users,
			Pages:   pages,
		}
		return response, nil
	}
}

func makeCreateCourse(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(CreateCourseRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}
		newUser := createCourseRequestToEntitiesCourse(req)
		userID, err := svc.CreateCourse(ctx, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		response := CreateCoursesResponse{
			ID: fmt.Sprintf("%d", userID),
		}
		return response, nil
	}
}

func makeUpdateCourse(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(UpdateCourseRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		newUser := updateCourseRequestToEntitiesCourse(req, intID)
		err = svc.UpdateCourse(ctx, intID, newUser)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return UpdateCourseResponse{}, nil
	}
}

func makeDeleteCourse(svc Service, log log.Logger) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(DeleteCourseRequest)
		if !ok {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		intID, err := strconv.Atoi(req.ID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "request", request)
			return nil, errors.New("could not decode")
		}

		err = svc.DeleteCourse(ctx, intID)
		if err != nil {
			level.Error(log).Log("message", "could not decode", "error", err)
			return nil, err
		}

		return DeleteCourseResponse{}, nil
	}
}
