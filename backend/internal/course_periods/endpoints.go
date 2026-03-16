package course_periods

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error)
	GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
	CreateCoursePeriod(ctx context.Context, period entities.CoursePeriod) (int64, error)
	UpdateCoursePeriod(ctx context.Context, periodID string, period entities.CoursePeriod) error
	DeleteCoursePeriod(ctx context.Context, periodID, userID string) error
	GetAnnouncement(ctx context.Context, announcementID string) (entities.Announcement, error)
	CreateAnnouncement(ctx context.Context, periodID string, announcement entities.Announcement) (int64, error)
	UpdateAnnouncement(ctx context.Context, announcementID string, announcement entities.Announcement) error
	DeleteAnnouncement(ctx context.Context, announcementID string) error
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

func (h *Handler) GetCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req GetCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	period, err := h.svc.GetCoursePeriod(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting course period", zap.String("id", req.ID), zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesCoursePeriodToGetCoursePeriodResponse(period))
}

func (h *Handler) GetCoursePeriods(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	courseID := c.Param("course_id")

	//nolint:errcheck
	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	periods, pages, err := h.svc.GetCoursePeriods(ctx, courseID, scope)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetCoursePeriodsResponse{
		Periods: EntitiesCoursePeriodsToGetCoursePeriodsResponse(periods),
		Pages:   pages,
	})
}

func (h *Handler) CreateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
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

func (h *Handler) UpdateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}

	updatedPeriod := updateCoursePeriodRequestToEntitiesCoursePeriod(req, userID)
	err := h.svc.UpdateCoursePeriod(ctx, updatedPeriod.ID, updatedPeriod)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}

	err := h.svc.DeleteCoursePeriod(ctx, req.ID, userID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

// Announcement handlers
func (h *Handler) GetAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req GetAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	announcement, err := h.svc.GetAnnouncement(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting announcement", zap.String("id", req.ID), zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesAnnouncementToGetAnnouncementResponse(announcement))
}

func (h *Handler) CreateAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	if err := c.Validate(&req); err != nil {
		h.log.Error("validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	announcement := createAnnouncementRequestToEntitiesAnnouncement(req)
	announcementID, err := h.svc.CreateAnnouncement(ctx, req.PeriodID, announcement)
	if err != nil {
		h.log.Error("could not create announcement", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateAnnouncementResponse{
		ID: fmt.Sprintf("%d", announcementID),
	})
}

func (h *Handler) UpdateAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	if err := c.Validate(&req); err != nil {
		h.log.Error("validation failed", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	announcement := updateAnnouncementRequestToEntitiesAnnouncement(req)
	err := h.svc.UpdateAnnouncement(ctx, req.ID, announcement)
	if err != nil {
		h.log.Error("could not update announcement", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	err := h.svc.DeleteAnnouncement(ctx, req.ID)
	if err != nil {
		h.log.Error("could not delete announcement", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}
