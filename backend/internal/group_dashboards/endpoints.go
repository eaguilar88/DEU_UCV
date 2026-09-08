package group_dashboards

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

func (h *Handler) RegisterDashboardProtectedEndpoints(g *echo.Group) {
	g.GET("/groups/:groupId/dashboard", h.GetGroupDashboard)
}

func (h *Handler) RegisterDashboardAdminEndpoints(g *echo.Group) {
	g.GET("/group_dashboards/faculty/:faculty", h.GetFacultyDashboard)
	g.GET("/group_dashboards/deu", h.GetDeuDashboard)
}

func (h *Handler) GetGroupDashboard(c echo.Context) error {
	var req GroupDashboardRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "identificador de grupo inválido")
	}
	
    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

	res, err := h.service.GetGroupDashboard(c.Request().Context(), req.GroupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error al consultar dashboard del grupo")
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetFacultyDashboard(c echo.Context) error {
	var req FacultyDashboardRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "nombre de facultad inválido")
	}

	res, err := h.service.GetFacultyDashboard(c.Request().Context(), req.Faculty)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error al consultar dashboard de facultad")
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetDeuDashboard(c echo.Context) error {
	res, err := h.service.GetDeuDashboard(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error al consultar dashboard DEU")
	}

	return c.JSON(http.StatusOK, res)
}
