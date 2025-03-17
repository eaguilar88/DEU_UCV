package course_periods

// TODO: Implement endpoints.go logic

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error)
	GetCoursePeriods(ctx context.Context, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
	CreateCoursePeriod(ctx context.Context, period entities.CoursePeriod) (int64, error)
	UpdateCoursePeriod(ctx context.Context, periodID int, period entities.CoursePeriod) error
	DeleteCoursePeriod(ctx context.Context, periodID int) error
}

type CoursePeriodEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeCoursePeriodEndpointsHandler(svc Service, log *zap.Logger) CoursePeriodEndpointsHandler {
	return CoursePeriodEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *CoursePeriodEndpointsHandler) GetCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetCoursePeriodRequest{ID: c.Param("id")}
	period, err := h.svc.GetCoursePeriod(ctx, req.ID)
	if err != nil {
		h.log.Error(fmt.Sprintf("error getting course period with ID: %s", req.ID), zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesCoursePeriodToGetCoursePeriodResponse(period))
}

func (h *CoursePeriodEndpointsHandler) GetCoursePeriods(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	scope.GetPageFromVars(c.QueryParam("page"))
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	periods, pages, err := h.svc.GetCoursePeriods(ctx, scope)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetCoursePeriodsResponse{
		Periods: EntitiesCoursePeriodsToGetCoursePeriodsResponse(periods),
		Pages:   pages,
	})
}

func (h *CoursePeriodEndpointsHandler) CreateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}
	userID, err := strconv.Atoi(c.Get("userID").(string))
	if err != nil {
		h.log.Error("could not decode userID from context", zap.Error(err))
		return echo.ErrBadRequest
	}

	periodID, err := h.svc.CreateCoursePeriod(ctx, createCoursePeriodRequestToEntitiesCoursePeriod(req, userID))
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateCoursePeriodResponse{
		ID: fmt.Sprintf("%d", periodID),
	})
}

func (h *CoursePeriodEndpointsHandler) UpdateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	userID, err := strconv.Atoi(c.Get("userID").(string))
	if err != nil {
		h.log.Error("could not decode userID from context", zap.Error(err))
		return echo.ErrBadRequest
	}

	updatedPeriod := updateCoursePeriodRequestToEntitiesCoursePeriod(req, userID)
	err = h.svc.UpdateCoursePeriod(ctx, updatedPeriod.ID, updatedPeriod)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *CoursePeriodEndpointsHandler) DeleteCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("invalid id", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	err = h.svc.DeleteCoursePeriod(ctx, intID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}
