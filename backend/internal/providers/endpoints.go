package providers

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

// Service defines the business logic operations for the providers domain.
type Service interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, providerCode string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider *entities.Provider) (int64, string, error)
	UpdateProvider(ctx context.Context, providerID string, provider *entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
}

// Handler holds the HTTP handler dependencies for the providers domain.
type Handler struct {
	svc Service
	log *zap.Logger
}

// NewHandler creates a new providers HTTP handler.
func NewHandler(svc Service, log *zap.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

func (h *Handler) GetProvider(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetProviderRequest{ID: c.Param("id")}
	provider, err := h.svc.GetProvider(ctx, req.ID)
	if err != nil {
		h.log.Error("error getting provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ProviderEntityToGetProviderResponse(provider))
}

func (h *Handler) GetProviders(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	req := GetProvidersRequest{
		PageScope: scope,
	}
	providers, pages, err := h.svc.GetProviders(ctx, req.PageScope)
	if err != nil {
		h.log.Error("error getting providers", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ProvidersEntityToGetProvidersResponse(providers, pages))
}

func (h *Handler) CreateProvider(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	provider, err := makeProviderFromRequest(c, userID, h.log)
	if err != nil {
		h.log.Error("error getting files", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	id, code, err := h.svc.CreateProvider(ctx, provider)
	if err != nil {
		h.log.Error("error creating provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, CreateProviderResponse{
		ID:   fmt.Sprintf("%d", id),
		Code: code,
	})
}

func (h *Handler) UpdateProvider(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	provider, err := makeProviderFromRequest(c, userID, h.log)
	if err != nil {
		h.log.Error("error getting files", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = h.svc.UpdateProvider(ctx, c.Param("id"), provider)
	if err != nil {
		h.log.Error("error updating provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) DeleteProvider(c echo.Context) error {
	ctx := c.Request().Context()
	err := h.svc.DeleteProvider(ctx, c.Param("id"))
	if err != nil {
		h.log.Error("error deleting provider", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func makeProviderFromRequest(c echo.Context, userID string, logger *zap.Logger) (*entities.Provider, error) {
	providerType := c.FormValue("tipo_proveedor")
	party := c.FormValue("tipo_persona")
	name := c.FormValue("nombre")
	bio := c.FormValue("bio")
	isInternal := c.FormValue("es_interno")

	ci, err := utils.GetFileFromForm(c, entities.ProviderFileTypeCI)
	if err != nil {
		logger.Error("error getting ci", zap.Error(err))
		return nil, errors.New("ci is required")
	}
	rif, err := utils.GetFileFromForm(c, entities.ProviderFileTypeRIF)
	if err != nil {
		logger.Error("error getting rif", zap.Error(err))
		return nil, errors.New("rif is required")
	}

	islr, err := utils.GetFileFromForm(c, entities.ProviderFileTypeISLR)
	if err != nil {
		logger.Error("error getting islr", zap.Error(err))
		return nil, errors.New("islr is required")
	}

	logo, err := utils.GetFileFromForm(c, entities.ProviderFileTypeLogo)
	if err != nil {
		logger.Error("error getting logo", zap.Error(err))
		return nil, errors.New("logo is required")
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
		Name:       name,
		Bio:        bio,
		Type:       entities.ProviderType(providerType),
		PartyType:  entities.ProviderPartyType(party),
		IsInternal: isInternal == "true",
		Files: entities.ProviderFiles{
			CI:   ci,
			RIF:  rif,
			ISLR: islr,
			Logo: logo,
		},
	}
	resumes := make([]*entities.File, 0, len(files["resumes"]))
	for _, resume := range files["resumes"] {
		file, err := resume.Open()
		if err != nil {
			logger.Error("error getting resume", zap.Error(err))
			return nil, errors.New("error getting resume")
		}
		resumes = append(resumes, &entities.File{
			Name:    resume.Filename,
			Body:    file,
			Purpose: entities.ProviderFileTypeResume,
		})
	}
	others := make([]*entities.File, 0, len(files["others"]))
	for _, other := range files["others"] {
		file, err := other.Open()
		if err != nil {
			logger.Error("error getting other", zap.Error(err))
			return nil, errors.New("error getting other")
		}
		others = append(others, &entities.File{
			Name:    other.Filename,
			Body:    file,
			Purpose: entities.ProviderFileTypeOther,
		})
	}

	provider.Files.Resumes = resumes
	provider.Files.Others = others

	return provider, nil
}
