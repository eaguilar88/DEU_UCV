package course_cycle_close_requests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/facultyscope"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	SubmitCloseRequest(ctx context.Context, request entities.CourseCycleCloseRequest, submittedByID string) (int64, error)
	ApproveCloseRequest(ctx context.Context, id, reviewerID string) error
	RejectCloseRequest(ctx context.Context, id, reviewerID, comments string) error
	GetCloseRequests(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseCycleCloseRequest, entities.PageScope, error)
	GetCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error)
	ParticipantsTemplate(ctx context.Context) ([]byte, error)
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

func (h *Handler) RegisterProtectedEndpoints(g *echo.Group) {
	g.POST("/course-cycle-close-requests", h.SubmitCloseRequest)
	g.GET("/course-cycle-close-requests/participants-template", h.GetParticipantsTemplate)
}

func (h *Handler) RegisterAdminEndpoints(g *echo.Group) {
	ccr := g.Group("/course-cycle-close-requests")
	ccr.GET("", h.GetCloseRequests)
	ccr.GET("/:id", h.GetCloseRequestByID)
	ccr.POST("/:id/approve", h.ApproveCloseRequest)
	ccr.POST("/:id/reject", h.RejectCloseRequest)
}

func (h *Handler) SubmitCloseRequest(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	request, err := toCloseRequestEntity(c)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	id, err := h.svc.SubmitCloseRequest(ctx, request, userID)
	if err != nil {
		var fileErr *ParticipantsFileError
		if errors.As(err, &fileErr) {
			return httperrors.NewBadRequest(fileErr.Error())
		}
		if errors.Is(err, ErrCloseRequestAlreadyPending) {
			return httperrors.NewConflict(ErrCloseRequestAlreadyPending.Error())
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// GetParticipantsTemplate downloads the XLSX template for the approved participants file.
func (h *Handler) GetParticipantsTemplate(c echo.Context) error {
	data, err := h.svc.ParticipantsTemplate(c.Request().Context())
	if err != nil {
		return httperrors.NewInternal(err)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", ParticipantsTemplateFileName))
	return c.Blob(http.StatusOK, ParticipantsTemplateContentType, data)
}

func (h *Handler) ApproveCloseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	if err := h.svc.ApproveCloseRequest(ctx, id, userID); err != nil {
		if errors.Is(err, ErrCycleCloseRequestNotFound) {
			return httperrors.NewNotFound("close request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RejectCloseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req RejectCloseRequestRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}

	if err := h.svc.RejectCloseRequest(ctx, id, userID, req.Comments); err != nil {
		if errors.Is(err, ErrCycleCloseRequestNotFound) {
			return httperrors.NewNotFound("close request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) GetCloseRequests(c echo.Context) error {
	ctx := c.Request().Context()

	faculty, err := facultyscope.ResolveOptional(c, c.QueryParam("faculty"))
	if err != nil {
		return httperrors.NewBadRequest("invalid faculty")
	}

	var pageScope entities.PageScope
	//nolint:errcheck
	pageScope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	pageScope.GetPerPageFromVars(c.QueryParam("per_page"))

	requests, resultScope, err := h.svc.GetCloseRequests(ctx, faculty, pageScope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	response := GetCloseRequestsResponse{
		Requests: make([]GetCloseRequestResponse, 0, len(requests)),
		Pages:    resultScope,
	}
	for _, r := range requests {
		response.Requests = append(response.Requests, closeRequestToResponse(r))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetCloseRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	req, err := h.svc.GetCloseRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrCycleCloseRequestNotFound) {
			return httperrors.NewNotFound("close request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, closeRequestToResponse(req))
}
