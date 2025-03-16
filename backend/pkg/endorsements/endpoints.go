package endorsements

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetEndorsement(ctx context.Context, endorsementID string) (entities.Endorsements, error)
	GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsements, entities.PageScope, error)
	CreateEndorsement(ctx context.Context, endorsement entities.Endorsements) (int64, error)
	UpdateEndorsement(ctx context.Context, endorsementID int, Endorsement entities.Endorsements) error
	DeleteEndorsement(ctx context.Context, endorsementID int) error
}

type EndorsementEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeEndorsementEndpointsHandler(svc Service, log *zap.Logger) EndorsementEndpointsHandler {
	return EndorsementEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *EndorsementEndpointsHandler) GetEndorsement(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetEndorsementRequest{ID: c.Param("id")}
	endorsement, err := h.svc.GetEndorsement(ctx, req.ID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, EntitiesEndorsementToGetEndorsementResponse(endorsement))
}

func (h *EndorsementEndpointsHandler) GetEndorsements(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	scope.GetPageFromVars(c.QueryParam("page"))
	scope.GetPerPageFromVars(c.QueryParam("per_page"))
	req := GetEndorsementsRequest{
		PageScope: scope,
	}
	endorsements, pages, err := h.svc.GetEndorsements(ctx, req.PageScope)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetEndorsementsResponse{
		Endorsements: EndorsementEntitiesToGetEndorsementsResponse(endorsements),
		Pages:        pages,
	})
}

func (h *EndorsementEndpointsHandler) CreateEndorsement(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateEndorsementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	userID, err := h.svc.CreateEndorsement(ctx, createEndorsementRequestToEntitiesEndorsement(req))
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateEndorsementResponse{
		ID: fmt.Sprintf("%d", userID),
	})
}

func (h *EndorsementEndpointsHandler) UpdateEndorsement(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateEndorsementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("invalid id", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	updatedEndorsement := updateEndorsementRequestToEntitiesEndorsement(req, intID)
	err = h.svc.UpdateEndorsement(ctx, intID, updatedEndorsement)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *EndorsementEndpointsHandler) DeleteEndorsement(c echo.Context) error {

	ctx := c.Request().Context()
	var req DeleteEndorsementRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("invalid id", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	err = h.svc.DeleteEndorsement(ctx, intID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}
