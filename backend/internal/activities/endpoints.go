package activities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetActivity(ctx context.Context, id string) (entities.Activity, error)
	GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error)
	CreateActivity(ctx context.Context, activity entities.Activity, files []*entities.File) (int64, error)
	UpdateActivity(ctx context.Context, id string, activity entities.Activity) error
	DeleteActivity(ctx context.Context, id string) error
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
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

	return c.JSON(http.StatusOK, activityToResponse(activity))
}

func (h *Handler) GetActivities(c echo.Context) error {
	ctx := c.Request().Context()

	var req GetActivitiesRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid query parameters")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("group_id is required")
	}

	scope := entities.PageScope{}
	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	filter := entities.ActivityFilter{
		GroupID:            req.GroupID,
		Date:               req.Date,
		ActualParticipants: req.ActualParticipants,
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

	reportFiles, err := getReportFiles(c)
	if err != nil {
		h.log.Error("failed to read report files", zap.Error(err))
		return httperrors.NewBadRequest("failed to read report files")
	}

	activity := createActivityEntityFromRequest(req)
	activityID, err := h.svc.CreateActivity(ctx, activity, reportFiles)
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

	err := h.svc.UpdateActivity(ctx, req.ID, updateActivityEntityFromRequest(req))
	if err != nil {
		if errors.Is(err, ErrActivityNotFound) {
			return httperrors.NewNotFound("activity not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
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

// getReportFiles reads all uploaded files from the "reporte" multipart field.
func getReportFiles(c echo.Context) ([]*entities.File, error) {
	if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	form := c.Request().MultipartForm
	if form == nil {
		return nil, nil
	}
	headers := form.File[entities.ActivityFileTypeReport]
	if len(headers) == 0 {
		return nil, nil
	}

	files := make([]*entities.File, 0, len(headers))
	for i, fh := range headers {
		body, err := fh.Open()
		if err != nil {
			return nil, err
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		files = append(files, &entities.File{
			Name:    strconv.Itoa(i) + "_" + entities.ActivityFileTypeReport + ext,
			Body:    body,
			Purpose: entities.ActivityFileTypeReport,
		})
	}
	return files, nil
}
