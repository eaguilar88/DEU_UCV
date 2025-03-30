package groups

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/pkg/endorsements"
	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/eaguilar88/deu/pkg/users"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// TODO: Implement endpoints.go logic
type Service interface {
	GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, error)
	UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID string) error
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

func EntitiesGroupsToGetGroupsResponse(groups []entities.ExtensionGroup) []GetGroupResponse {
	var res []GetGroupResponse
	for _, group := range groups {
		res = append(res, EntitiesGroupToGetGroupResponse(group))
	}
	return res
}

func EntitiesGroupToGetGroupResponse(group entities.ExtensionGroup) GetGroupResponse {
	owner := users.UserEntityToGetUserResponse(group.Owner)
	endorsement := endorsements.EntitiesEndorsementToGetEndorsementResponse(group.Endorsement)
	return GetGroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		Owner:       &owner,
		Endorsement: &endorsement,
		Objective:   group.Objective,
		Location:    group.Location,
		Active:      group.Active,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
