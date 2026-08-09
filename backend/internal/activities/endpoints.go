package activities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	clientDateFormat  = "02-01-2006" // DD-MM-YYYY, matches internal/users/decoders.go's client-facing format
	storageDateFormat = "2006-01-02"
)

type Service interface {
	GetActivity(ctx context.Context, id string) (entities.Activity, error)
	GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error)
	GetGroupDashboardSummary(ctx context.Context, groupID string) (entities.GroupDashboardSummary, error)
	CreateActivity(ctx context.Context, activity entities.Activity) (int64, error)
	UpdateActivity(ctx context.Context, id string, activity entities.Activity) error
	ToggleReportCheck(ctx context.Context, id string, checked bool) error
	ToggleFeature(ctx context.Context, id string, featured bool) error
	DeleteActivity(ctx context.Context, id string) error
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// RegisterActivityAdminEndpoints registra las rutas restringidas a administradores
func (h *Handler) RegisterActivityAdminEndpoints(g *echo.Group) {
	g.PATCH("/activities/report-check", h.ToggleReportCheck)
}

func (h *Handler) GetActivity(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("activity id is required")
	}

	activity, err := h.svc.GetActivity(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, activityToResponse(activity, true))
}

func (h *Handler) GetActivities(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetActivitiesRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid query parameters")
	}

	startDate, endDate, err := h.buildDateFilter(c)
	if err != nil {
		return err
	}

	scope := entities.PageScope{}
	if !req.DisablePaging {
		//nolint:errcheck
		scope.GetPageFromVars(c.QueryParam("page"))
		//nolint:errcheck
		scope.GetPerPageFromVars(c.QueryParam("per_page"))
	}

	filter := entities.ActivityFilter{
		GroupID:               req.GroupID,
		NameSearch:            req.NameSearch,
		StartDate:             startDate,
		EndDate:               endDate,
		ActualParticipants:    req.ActualParticipants,
		HasActualParticipants: req.HasActualParticipants,
		IsFeatured:            req.IsFeatured,
		ReportChecked:         req.ReportChecked,
		Order:                 req.Order,
		DisablePaging:         req.DisablePaging,
	}

	list, pageScope, err := h.svc.GetActivities(ctx, filter, scope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetActivitiesResponse{
		Activities: activitiesToResponse(list),
		PageScope:  pageScope,
	})
}

func (h *Handler) GetGroupDashboardSummary(c echo.Context) error {
	ctx := c.Request().Context()
	groupID := c.Param("groupId")
	if groupID == "" {
		return httperrors.NewBadRequest("group id is required")
	}
	summary, err := h.svc.GetGroupDashboardSummary(ctx, groupID)
	if err != nil {
		return httperrors.NewInternal(err)
	}
	return c.JSON(http.StatusOK, summary)
}

// buildDateFilter parses and validates the start_date/end_date query params,
// returning them in storage format (YYYY-MM-DD), or an httperrors.CustomError.
func (h *Handler) buildDateFilter(c echo.Context) (string, string, error) {
	startStr := c.QueryParam("start_date")
	endStr := c.QueryParam("end_date")

	if startStr == "" && endStr != "" {
		return "", "", httperrors.NewBadRequest("start_date is required when end_date is provided")
	}
	if startStr == "" {
		return "", "", nil
	}

	start, err := time.Parse(clientDateFormat, startStr)
	if err != nil {
		return "", "", httperrors.NewBadRequest("start_date must be in DD-MM-YYYY format")
	}

	end := time.Now()
	if endStr != "" {
		end, err = time.Parse(clientDateFormat, endStr)
		if err != nil {
			return "", "", httperrors.NewBadRequest("end_date must be in DD-MM-YYYY format")
		}
	}

	if start.After(end) {
		return "", "", httperrors.NewBadRequest("start_date must not be after end_date")
	}

	return start.Format(storageDateFormat), end.Format(storageDateFormat), nil
}

func (h *Handler) CreateActivity(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req CreateActivityRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	//coverImage, err := utils.GetFileFrom(c, entities.ActivityFileTypeCoverImage)
	//if err != nil {
	//	return httperrors.NewBadRequest("cubierta is required")
	//}

	activity := createActivityEntityFromRequest(req)

	if coverImage, err := utils.GetFileFrom(c, entities.ActivityFileTypeCoverImage); err == nil {
		activity.CoverImage = coverImage
	}
	if participantList, err := utils.GetFileFrom(c, entities.ActivityFileTypeListParticipants); err == nil {
		activity.ParticipantList = participantList
	}

	activityID, err := h.svc.CreateActivity(ctx, activity)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateActivityResponse{
		ID: fmt.Sprintf("%d", activityID),
	})
}

func (h *Handler) UpdateActivity(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req UpdateActivityRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	req.ID = c.Param("id")
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	activity := updateActivityEntityFromRequest(req)

	if coverImage, err := utils.GetFileFrom(c, entities.ActivityFileTypeCoverImage); err == nil {
		activity.CoverImage = coverImage
	}
	if participantList, err := utils.GetFileFrom(c, entities.ActivityFileTypeListParticipants); err == nil {
		activity.ParticipantList = participantList
	}

	err := h.svc.UpdateActivity(ctx, req.ID, activity)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) ToggleReportCheck(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req ToggleReportCheckRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	if err := h.svc.ToggleReportCheck(ctx, req.ID, req.ReportChecked); err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "status updated successfully"})
}

func (h *Handler) ToggleFeature(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req ToggleFeatureRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	if err := h.svc.ToggleFeature(ctx, req.ID, req.IsFeatured); err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		if errors.Is(err, ErrMaxFeaturedLimitReached) {
			return httperrors.NewBadRequest("Solo se pueden destacar un máximo de 4 actividades por grupo")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "status updated successfully"})
}

func (h *Handler) DeleteActivity(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("activity id is required")
	}

	err := h.svc.DeleteActivity(ctx, id)
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusNoContent, nil)
}
