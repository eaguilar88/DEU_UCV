package groups

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TODO: Implement endpoints.go logic
type Service interface {
	GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup, userID string) (int64, error)
	UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID, userID string) error
}

type GroupEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeGroupEndpointsHandler(svc Service, log *zap.Logger) GroupEndpointsHandler {
	return GroupEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *GroupEndpointsHandler) GetGroup(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetGroupRequest{ID: c.Param("id")}
	group, err := h.svc.GetGroup(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesGroupToGetGroupResponse(group))
}

func (h *GroupEndpointsHandler) GetGroups(c echo.Context) error {
	ctx := c.Request().Context()
	var req GetGroupsRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error binding request", zap.Error(err))
		return echo.ErrBadRequest
	}
	groups, pageScope, err := h.svc.GetGroups(ctx, entities.PageScope{
		Page:    req.Page,
		PerPage: req.PerPage,
	})
	if err != nil {
		h.log.Error("error getting groups", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetGroupsResponse{
		Groups:    EntitiesGroupsToGetGroupsResponse(groups),
		PageScope: pageScope,
	})
}

func (h *GroupEndpointsHandler) CreateGroup(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error binding request", zap.Error(err))
		return echo.ErrBadRequest
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}
	req.OwnerID = userID
	if err := c.Validate(req); err != nil {
		h.log.Error("error validating request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	groupID, err := h.svc.CreateGroup(ctx, createGroupEntityFromRequest(req), userID)
	if err != nil {
		h.log.Error("error creating group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateGroupResponse{
		ID: fmt.Sprintf("%d", groupID),
	})
}

func (h *GroupEndpointsHandler) UpdateGroup(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateGroupRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error binding request", zap.Error(err))
		return echo.ErrBadRequest
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}
	req.OwnerID = userID
	if err := c.Validate(req); err != nil {
		h.log.Error("error validating request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.svc.UpdateGroup(ctx, req.ID, updateGroupEntityFromRequest(req))
	if err != nil {
		h.log.Error("error updating group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusAccepted, nil)
}
func (h *GroupEndpointsHandler) DeleteGroup(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteGroupRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("error binding request", zap.Error(err))
		return echo.ErrBadRequest
	}
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}
	req.OwnerID = userID
	if err := c.Validate(req); err != nil {
		h.log.Error("error validating request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.svc.DeleteGroup(ctx, req.ID, req.OwnerID)
	if err != nil {
		h.log.Error("error deleting group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusNoContent, nil)
}
