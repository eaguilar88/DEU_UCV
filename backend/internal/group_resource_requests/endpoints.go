package group_resource_requests

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error)
	ApproveGroupResourceRequest(ctx context.Context, reqID string) error
	RejectGroupResourceRequest(ctx context.Context, reqID string) error
	GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error)
	GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error)
}

type Handler struct {
	svc Service
	log *zap.Logger
}

func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) RegisterGroupResourceRequestEndpoints(g *echo.Group) {
	gr := g.Group("/group-resource-requests")
	gr.POST("", h.CreateGroupResourceRequest)
}

func (h *Handler) RegisterGroupResourceRequestAdminEndpoints(g *echo.Group) {
	gr := g.Group("/group-resource-requests")
	gr.GET("", h.GetGroupResourceRequestsByFaculty)
	gr.GET("/:id", h.GetGroupResourceRequestByID)
	gr.POST("/:id/approve", h.ApproveGroupResourceRequest)
	gr.POST("/:id/reject", h.RejectGroupResourceRequest)
}

type createGroupResourceRequestInput struct {
	GroupID string `json:"grupo_id"  validate:"required,min=1"`
	Type    string `json:"tipo"      validate:"required,min=1"`
	Content string `json:"contenido"`
}

func (h *Handler) CreateGroupResourceRequest(c echo.Context) error {
	ctx := c.Request().Context()
	var input createGroupResourceRequestInput
	if err := c.Bind(&input); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(input); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	id, err := h.svc.CreateGroupResourceRequest(ctx, entities.GroupResourceRequest{
		GroupID: input.GroupID,
		Type:    input.Type,
		Content: input.Content,
		Status:  "under_review",
	})
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateGroupResourceRequestResponse{
		ID: fmt.Sprintf("%d", id),
	})
}

func (h *Handler) ApproveGroupResourceRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}
	if err := h.svc.ApproveGroupResourceRequest(ctx, reqID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return httperrors.NewNotFound("group resource request not found")
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RejectGroupResourceRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}
	if err := h.svc.RejectGroupResourceRequest(ctx, reqID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return httperrors.NewNotFound("group resource request not found")
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) GetGroupResourceRequestsByFaculty(c echo.Context) error {
	ctx := c.Request().Context()
	faculty, err := entities.FromString(c.QueryParam("faculty"))
	if err != nil {
		return httperrors.NewBadRequest("invalid faculty")
	}

	var pageScope entities.PageScope
	//nolint:errcheck
	pageScope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	pageScope.GetPerPageFromVars(c.QueryParam("pageSize"))

	reqs, resultScope, err := h.svc.GetGroupResourceRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	response := GetGroupResourceRequestsResponse{
		Requests: make([]GetGroupResourceRequestResponse, 0, len(reqs)),
		Pages:    resultScope,
	}
	for _, req := range reqs {
		response.Requests = append(response.Requests, toResponse(req))
	}
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetGroupResourceRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}
	req, err := h.svc.GetGroupResourceRequestByID(ctx, reqID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return httperrors.NewNotFound("group resource request not found")
		}
		return httperrors.NewInternal(err)
	}
	return c.JSON(http.StatusOK, toResponse(req))
}

func toResponse(r entities.GroupResourceRequest) GetGroupResourceRequestResponse {
	return GetGroupResourceRequestResponse{
		ID:        r.ID,
		GroupID:   r.GroupID,
		Type:      r.Type,
		Content:   r.Content,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
