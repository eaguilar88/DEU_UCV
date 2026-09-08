package group_analytics

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterAnalyticsProtectedEndpoints(g *echo.Group) {
	g.POST("/analytics", h.GetAnalytics)
}

func (h *Handler) RegisterAnalyticsAdminEndpoints(g *echo.Group) {
	g.POST("/analytics", h.GetAnalytics)
}

func (h *Handler) GetAnalytics(c echo.Context) error {
	var req AnalyticsFilterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "datos de consulta inválidos")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.service.GetAnalytics(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error al generar analíticas de grupos")
	}

	return c.JSON(http.StatusOK, res)
}
