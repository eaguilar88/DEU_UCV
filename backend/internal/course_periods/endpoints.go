package course_periods

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
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
		return errors.NewBadRequest("invalid request")
	}

	period, err := h.svc.GetCoursePeriod(ctx, req.ID)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, periodToResponse(period))
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
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetCoursePeriodsResponse{
		Periods: periodsToResponse(periods),
		Pages:   pages,
	})
}

func (h *Handler) CreateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		return errors.NewUnauthorized("authentication required")
	}

	periodID, err := h.svc.CreateCoursePeriod(ctx, toPeriodEntity(req, userID))
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateCoursePeriodResponse{
		ID: fmt.Sprintf("%d", periodID),
	})
}

func (h *Handler) UpdateCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return errors.NewUnauthorized("authentication required")
	}

	updatedPeriod := toPeriodUpdateEntity(req, userID)
	err := h.svc.UpdateCoursePeriod(ctx, updatedPeriod.ID, updatedPeriod)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteCoursePeriod(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCoursePeriodRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		return errors.NewUnauthorized("authentication required")
	}

	err := h.svc.DeleteCoursePeriod(ctx, req.ID, userID)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

// Announcement handlers
func (h *Handler) GetAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req GetAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request")
	}

	announcement, err := h.svc.GetAnnouncement(ctx, req.ID)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, announcementToResponse(announcement))
}

func (h *Handler) CreateAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewBadRequest("validation failed")
	}

	announcement := toAnnouncementEntity(req)
	announcementID, err := h.svc.CreateAnnouncement(ctx, req.PeriodID, announcement)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateAnnouncementResponse{
		ID: fmt.Sprintf("%d", announcementID),
	})
}

func (h *Handler) UpdateAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return errors.NewBadRequest("validation failed")
	}

	announcement := toAnnouncementUpdateEntity(req)
	err := h.svc.UpdateAnnouncement(ctx, req.ID, announcement)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteAnnouncement(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	err := h.svc.DeleteAnnouncement(ctx, req.ID)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}
