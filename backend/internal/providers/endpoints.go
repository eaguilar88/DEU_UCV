package providers

import (
	"context"
	"errors"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, providerCode string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider *entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider *entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
}

// TODO: Implement endpoints.go logic
type ProviderEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeProviderEndpointsHandler(svc Service, log *zap.Logger) ProviderEndpointsHandler {
	return ProviderEndpointsHandler{
		svc: svc,
		log: log,
	}
}
func (h *ProviderEndpointsHandler) GetProvider(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetProviderRequest{ID: c.Param("id")}
	provider, err := h.svc.GetProvider(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ProviderEntityToGetProviderResponse(provider))
}
func (h *ProviderEndpointsHandler) GetProviders(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	scope.GetPageFromVars(c.QueryParam("page"))
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	req := GetProvidersRequest{
		PageScope: scope,
	}
	providers, pages, err := h.svc.GetProviders(ctx, req.PageScope)
	if err != nil {
		h.log.Error("error getting providers", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, ProvidersEntityToGetProvidersResponse(providers, pages))
}
func (h *ProviderEndpointsHandler) CreateProvider(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	provider, err := makeProviderFromRequest(c, userID)
	if err != nil {
		h.log.Error("error getting files", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	id, err := h.svc.CreateProvider(ctx, provider)
	if err != nil {
		h.log.Error("error creating provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, id)
}
func (h *ProviderEndpointsHandler) UpdateProvider(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	provider, err := makeProviderFromRequest(c, userID)
	if err != nil {
		h.log.Error("error getting files", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.svc.UpdateProvider(ctx, c.Param("id"), provider)
	if err != nil {
		h.log.Error("error updating provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusAccepted)
}
func (h *ProviderEndpointsHandler) DeleteProvider(c echo.Context) error {
	ctx := c.Request().Context()
	err := h.svc.DeleteProvider(ctx, c.Param("id"))
	if err != nil {
		h.log.Error("error deleting provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func makeProviderFromRequest(c echo.Context, userID string) (*entities.Provider, error) {
	providerType := c.FormValue("provider_type")
	isInternal := c.FormValue("is_internal")
	ci, err := c.FormFile("ci")
	if err != nil {
		return nil, errors.New("ci is required")
	}

	rif, err := c.FormFile("rif")
	if err != nil {
		return nil, errors.New("rif is required")
	}

	islr, err := c.FormFile("islr")
	if err != nil {
		return nil, errors.New("islr is required")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return nil, errors.New("error parsing form")
	}

	files := form.File
	if len(files["resumes"]) == 0 {
		return nil, errors.New("at least one resume is required")
	}

	provider := &entities.Provider{
		User: entities.User{
			ID: userID,
		},
		Type:       entities.ProviderType(providerType),
		IsInternal: isInternal == "true",
		CI:         ci,
		RIF:        rif,
		ISLR:       islr,
	}

	provider.Resumes = append(provider.Resumes, files["resumes"]...)
	provider.Others = append(provider.Others, files["others"]...)

	return provider, nil
}
