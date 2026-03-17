package group_requests

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
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

func (h *Handler) RegisterGroupRequestAdminEndpoints(g *echo.Group) {
	gr := g.Group("/group-requests")
	gr.GET("", h.GetGroupRequestsByFaculty)
	gr.GET("/:id", h.GetGroupRequestByID)
	gr.POST("/:id/approve", h.ApproveGroupRequest)
	gr.POST("/:id/reject", h.RejectGroupRequest)
}

func (h *Handler) ApproveGroupRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return errors.NewBadRequest("request ID is required")
	}

	if err := h.svc.ApproveGroupRequest(ctx, reqID); err != nil {
		return errors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RejectGroupRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return errors.NewBadRequest("request ID is required")
	}

	if err := h.svc.RejectGroupRequest(ctx, reqID); err != nil {
		return errors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) GetGroupRequestsByFaculty(c echo.Context) error {
	ctx := c.Request().Context()
	faculty, err := entities.FromString(c.QueryParam("faculty"))
	if err != nil {
		return errors.NewBadRequest("invalid faculty")
	}

	var pageScope entities.PageScope
	if err := pageScope.GetPageFromVars(c.QueryParam("page")); err != nil {
		return errors.NewBadRequest("invalid page number")
	}

	if err := pageScope.GetPerPageFromVars(c.QueryParam("pageSize")); err != nil {
		return errors.NewBadRequest("invalid page size")
	}

	requests, resultScope, err := h.svc.GetGroupRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		return errors.NewInternal(err)
	}

	response := GetGroupRequestsResponse{
		Requests: make([]GetGroupRequestResponse, 0, len(requests)),
		Pages:    resultScope,
	}

	for _, req := range requests {
		response.Requests = append(response.Requests, GetGroupRequestResponse{
			ID:        req.ID,
			GroupID:   req.GroupID,
			Comments:  req.Comments,
			Status:    string(req.Status),
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetGroupRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return errors.NewBadRequest("request ID is required")
	}

	req, err := h.svc.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return errors.NewInternal(err)
	}

	response := GetGroupRequestResponse{
		ID:        reqID,
		GroupID:   req.GroupID,
		Comments:  req.Comments,
		Status:    string(req.Status),
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
}
