package groups

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, string, error)
	UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID, userID string) error
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

func (h *Handler) GetGroup(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetGroupRequest{ID: c.Param("id")}
	group, err := h.svc.GetGroup(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesGroupToGetGroupResponse(group))
}

func (h *Handler) GetGroups(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	groups, pageScope, err := h.svc.GetGroups(ctx, scope)
	if err != nil {
		h.log.Error("error getting groups", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetGroupsResponse{
		Groups:    EntitiesGroupsToGetGroupsResponse(groups),
		PageScope: pageScope,
	})
}

func (h *Handler) CreateGroup(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.ErrBadRequest
	}

	req, err := makeGroupRequestFromContext(c, userID, h.log)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	groupID, providerCode, err := h.svc.CreateGroup(ctx, req)
	if err != nil {
		h.log.Error("error creating group", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateGroupResponse{
		ID:   fmt.Sprintf("%d", groupID),
		Code: providerCode,
	})
}

func (h *Handler) UpdateGroup(c echo.Context) error {
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

func (h *Handler) DeleteGroup(c echo.Context) error {
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

func makeGroupRequestFromContext(c echo.Context, userID string, log *zap.Logger) (entities.ExtensionGroup, error) {
	var req CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		log.Error("error binding request", zap.Error(err))
		return entities.ExtensionGroup{}, err
	}

	if err := c.Validate(req); err != nil {
		log.Error("error validating request", zap.Error(err))
		return entities.ExtensionGroup{}, err
	}

	logo, err := utils.GetFileFromForm(c, entities.GroupFileTypeLogo)
	if err != nil {
		log.Error("error getting logo", zap.Error(err))
		return entities.ExtensionGroup{}, errors.New("logo is required")
	}

	// fp, err := utils.GetFileFromForm(c, entities.GroupFileTypeFinancingPlan)
	// if err != nil {
	// 	log.Error("error getting financing plan", zap.Error(err))
	// 	return entities.ExtensionGroup{}, errors.New("financing plan is required")
	// }

	// gp, err := utils.GetFileFromForm(c, entities.GroupFileTypeGroupProject)
	// if err != nil {
	// 	log.Error("error getting group project", zap.Error(err))
	// 	return entities.ExtensionGroup{}, errors.New("group project is required")
	// }

	group := createGroupEntityFromRequest(req, userID, req.Faculty)

	group.Files = &entities.GroupFiles{
		Logo: logo,
		// FinancingPlan: fp,
		// GroupProject:  gp,
	}
	return group, nil
}
