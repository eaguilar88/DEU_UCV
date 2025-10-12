package course_requests

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetCourseRequest(ctx context.Context, courseRequestID string) (entities.CourseRequest, error)
	GetCourseRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	CreateCourseRequest(ctx context.Context, courseRequest entities.CourseRequest) (int64, error)
	UpdateCourseRequest(ctx context.Context, courseRequest entities.CourseRequest) error
	DeleteCourseRequest(ctx context.Context, courseRequestID string) error
}

type CourseRequestEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeCourseRequestEndpointsHandler(svc Service, log *zap.Logger) CourseRequestEndpointsHandler {
	return CourseRequestEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *CourseRequestEndpointsHandler) GetCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetCourseRequestRequest{ID: c.Param("id")}
	courseRequest, err := h.svc.GetCourseRequest(ctx, req.ID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, EntitiesCourseRequestToGetCourseRequestResponse(courseRequest))
}

func (h *CourseRequestEndpointsHandler) GetCourseRequests(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))
	req := GetCourseRequestsRequest{
		PageScope: scope,
	}
	courseRequests, pages, err := h.svc.GetCourseRequests(ctx, req.PageScope)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetCourseRequestsResponse{
		CourseRequests: CourseRequestEntitiesToGetCourseRequestsResponse(courseRequests),
		Pages:          pages,
	})
}

func (h *CourseRequestEndpointsHandler) CreateCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	userID, err := h.svc.CreateCourseRequest(ctx, createCourseRequestRequestToEntitiesCourseRequest(req))
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateCourseRequestResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

func (h *CourseRequestEndpointsHandler) UpdateCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	updatedCourseRequest := updateCourseRequestRequestToEntitiesCourseRequest(req)
	err := h.svc.UpdateCourseRequest(ctx, updatedCourseRequest)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *CourseRequestEndpointsHandler) DeleteCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID := c.Param("id")
	err := h.svc.DeleteCourseRequest(ctx, intID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}
