package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Service defines the business logic operations for the providers domain.
type Service interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, providerCode string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope, filters entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider *entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider *entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
	UploadProviderDocuments(ctx context.Context, userID string, intentionLetter, commitmentLetter *entities.File) error
	ApproveProvider(ctx context.Context, providerID, userID string) error
	RejectProvider(ctx context.Context, providerID string) error
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

func (h *Handler) RegisterProviderAdminEndpoints(g *echo.Group) {
	gr := g.Group("/providers")
	gr.GET("", h.GetProviders)
	gr.POST("/:id/approve", h.ApproveProvider)
	gr.POST("/:id/reject", h.RejectProvider)
}

func (h *Handler) ApproveProvider(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}
	if err := h.svc.ApproveProvider(ctx, c.Param("id"), userID); err != nil {
		return httperrors.NewInternal(err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) RejectProvider(c echo.Context) error {
	ctx := c.Request().Context()
	if err := h.svc.RejectProvider(ctx, c.Param("id")); err != nil {
		return httperrors.NewInternal(err)
	}
	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) GetProvider(c echo.Context) error {
	ctx := c.Request().Context()
	provider, err := h.svc.GetProvider(ctx, c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return httperrors.NewNotFound("provider not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, providerToResponse(provider))
}

func (h *Handler) GetProviders(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	filters := entities.ProviderFilters{
		Type: entities.ProviderType(c.QueryParam("type")),
	}

	if faculty, ok := c.Get("faculty").(string); ok {
		// the logged user has a faculty on its claims
		filters.Faculty = entities.Faculty(faculty)
	} else {
		if v := c.QueryParam("faculty"); v != "" {
			f, err := entities.FromString(v)
			if err == nil {
				filters.Faculty = f
			}
		}
	}

	if roles, ok := c.Get("roles").([]string); ok && isAdmin(roles) {
		filters.IsAdmin = true
		filters.PartyType = entities.ProviderPartyType(c.QueryParam("party_type"))
		filters.ProfitType = entities.ProviderProfitType(c.QueryParam("profit_type"))
		filters.Code = c.QueryParam("code")
		filters.CreatedAtFrom = c.QueryParam("created_at_from")
		filters.CreatedAtTo = c.QueryParam("created_at_to")
		if v := c.QueryParam("is_internal"); v != "" {
			b := v == "true"
			filters.IsInternal = &b
		}
		if v := c.QueryParam("status"); v != "" {
			filters.Status = entities.ProviderStatus(v)
		}
	}

	providers, pages, err := h.svc.GetProviders(ctx, scope, filters)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, providersToResponse(providers, pages))
}

func (h *Handler) CreateProvider(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	provider, err := makeProviderFromRequest(c, userID, h.log)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	id, err := h.svc.CreateProvider(ctx, provider)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateProviderResponse{
		ID: fmt.Sprintf("%d", id),
	})
}

func (h *Handler) UpdateProvider(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	provider, err := makeProviderFromRequest(c, userID, h.log)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	err = h.svc.UpdateProvider(ctx, c.Param("id"), provider)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return httperrors.NewNotFound("provider not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) DeleteProvider(c echo.Context) error {
	ctx := c.Request().Context()
	err := h.svc.DeleteProvider(ctx, c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return httperrors.NewNotFound("provider not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) UploadProviderDocuments(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	commitmentLetter, err := utils.GetFileFrom(c, "carta_compromiso")
	if err != nil {
		return httperrors.NewBadRequest("carta_compromiso es requerida")
	}

	var intentionLetter *entities.File
	if il, err := utils.GetFileFrom(c, "carta_intencion"); err == nil {
		intentionLetter = il
	}

	if err := h.svc.UploadProviderDocuments(c.Request().Context(), userID, intentionLetter, commitmentLetter); err != nil {
		if errors.Is(err, ErrNoIntentionLetter) {
			return httperrors.NewBadRequest(err.Error())
		}
		if errors.Is(err, ErrProviderNotFound) {
			return httperrors.NewNotFound("provider not found")
		}
		return httperrors.NewInternal(err)
	}
	return c.NoContent(http.StatusCreated)
}

func makeProviderFromRequest(c echo.Context, userID string, logger *zap.Logger) (*entities.Provider, error) {
	providerType := c.FormValue("tipo_proveedor")
	party := c.FormValue("tipo_persona")
	profitType := c.FormValue("tipo_lucro")
	name := c.FormValue("nombre")
	bio := c.FormValue("bio")
	isInternal := c.FormValue("es_interno")
	faculty := c.FormValue("facultad")

	if providerType != string(entities.CourseProviderType) && providerType != string(entities.GroupProviderType) {
		return nil, errors.New("tipo_proveedor must be 'courses' or 'groups'")
	}

	if party != string(entities.PartyTypeNatural) && party != string(entities.PartyTypeJuridical) {
		return nil, errors.New("tipo_persona must be 'natural' or 'juridical'")
	}

	if profitType != string(entities.ProfitTypeLucrativo) && profitType != string(entities.ProfitTypeNoLucrativo) {
		return nil, errors.New("tipo_lucro must be 'lucrativo' or 'no_lucrativo'")
	}

	ci, err := utils.GetFileFrom(c, entities.ProviderFileTypeCI)
	if err != nil {
		logger.Error("error getting ci", zap.Error(err))
		return nil, errors.New("ci is required")
	}
	rif, err := utils.GetFileFrom(c, entities.ProviderFileTypeRIF)
	if err != nil {
		logger.Error("error getting rif", zap.Error(err))
		return nil, errors.New("rif is required")
	}

	islr, err := utils.GetFileFrom(c, entities.ProviderFileTypeISLR)
	if err != nil {
		logger.Error("error getting islr", zap.Error(err))
		return nil, errors.New("islr is required")
	}

	logo, err := utils.GetFileFrom(c, entities.ProviderFileTypeLogo)
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
		ProfitType: entities.ProviderProfitType(profitType),
		IsInternal: isInternal == "true",
		Faculty:    entities.Faculty(faculty),
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

func isAdmin(roles []string) bool {
	return slices.Contains(roles, "faculty_admin") ||
		slices.Contains(roles, "deu_admin") ||
		slices.Contains(roles, "group_admin")
}
