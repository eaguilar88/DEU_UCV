package provider_requests

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	ApproveProviderRequest(ctx context.Context, id, reviewerID string) error
	RejectProviderRequest(ctx context.Context, id, reviewerID, comments string) error
	GetProviderRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.ProviderRequest, entities.PageScope, error)
	GetProviderRequestByID(ctx context.Context, id string) (entities.ProviderRequest, error)
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

func (h *Handler) RegisterProviderRequestAdminEndpoints(g *echo.Group) {
	pr := g.Group("/provider-requests")
	pr.GET("", h.GetProviderRequests)
	pr.GET("/:id", h.GetProviderRequestByID)
	pr.POST("/:id/approve", h.ApproveProviderRequest)
	pr.POST("/:id/reject", h.RejectProviderRequest)
}

func (h *Handler) ApproveProviderRequest(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	if err := h.svc.ApproveProviderRequest(ctx, id, userID); err != nil {
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RejectProviderRequest(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req RejectProviderRequestRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}

	if err := h.svc.RejectProviderRequest(ctx, id, userID, req.Comments); err != nil {
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) GetProviderRequests(c echo.Context) error {
	ctx := c.Request().Context()

	var pageScope entities.PageScope
	//nolint:errcheck
	pageScope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	pageScope.GetPerPageFromVars(c.QueryParam("per_page"))

	requests, resultScope, err := h.svc.GetProviderRequests(ctx, pageScope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	response := GetProviderRequestsResponse{
		Requests: make([]GetProviderRequestResponse, 0, len(requests)),
		Pages:    resultScope,
	}
	for _, r := range requests {
		response.Requests = append(response.Requests, providerRequestToResponse(r))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetProviderRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	if id == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	req, err := h.svc.GetProviderRequestByID(ctx, id)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, providerRequestToResponse(req))
}
