package groups

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/utils"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	defaultRandomGroupsLimit = 3
	maxRandomGroupsLimit     = 5
)

type Service interface {
	GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error)
	GetGroupsSimple(ctx context.Context) ([]entities.ExtensionGroup, error)
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
		return mapGroupError(err)
	}

	return c.JSON(http.StatusOK, groupToResponse(group))
}

// userIDFromContext extracts the authenticated user ID set by the auth middleware.
func userIDFromContext(c echo.Context) (string, error) {
	userID, ok := c.Get("userID").(string)
	if !ok {
		return "", httperrors.NewUnauthorized("authentication required")
	}
	return userID, nil
}

// mapGroupError translates a service-layer error into the appropriate HTTP error.
func mapGroupError(err error) error {
	if errors.Is(err, ErrGroupNotFound) {
		return httperrors.NewNotFound("group not found")
	}
	return httperrors.NewInternal(err)
}

func (h *Handler) GetGroups(c echo.Context) error {
	ctx := c.Request().Context()

	// Si se solicita la lista simplificada para selectores
	if c.QueryParam("simplelist") == "true" {
		groups, err := h.svc.GetGroupsSimple(ctx)
		if err != nil {
			return httperrors.NewInternal(err)
		}

		type SimpleGroupResponse struct {
			ID     string `json:"id"`
			Nombre string `json:"nombre"`
		}

		simpleGroups := make([]SimpleGroupResponse, 0, len(groups))
		for _, g := range groups {
			simpleGroups = append(simpleGroups, SimpleGroupResponse{
				ID:     g.ID,
				Nombre: g.Name,
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"grupos": simpleGroups,
		})
	}

	if c.QueryParam("random") == "true" {
		limit := defaultRandomGroupsLimit
		if l, err := strconv.Atoi(c.QueryParam("limit")); err == nil && l > 0 {
			limit = l
		}
		if limit > maxRandomGroupsLimit {
			limit = maxRandomGroupsLimit
		}

		groups, err := h.svc.GetRandomActiveGroups(ctx, limit)
		if err != nil {
			return httperrors.NewInternal(err)
		}

		return c.JSON(http.StatusOK, GetGroupsResponse{
			Groups: groupsToResponse(groups),
		})
	}

	filter, err := h.buildFilters(c)
	if err != nil {
		return err
	}

	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))

	groups, pageScope, err := h.svc.GetGroups(ctx, filter, scope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetGroupsResponse{
		Groups:    groupsToResponse(groups),
		PageScope: pageScope,
	})
}

// buildFilters parses and validates the faculty, type, active, and deleted
// query params into an entities.GroupFilter.
func (h *Handler) buildFilters(c echo.Context) (entities.GroupFilter, error) {
	filter := entities.GroupFilter{}

	if facultyStr := c.QueryParam("faculty"); facultyStr != "" {
		faculty, err := ValidateFaculty(facultyStr)
		if err != nil {
			return entities.GroupFilter{}, httperrors.NewBadRequest(err.Error())
		}
		filter.Faculty = faculty
	}

	if typeStr := c.QueryParam("type"); typeStr != "" {
		groupType := entities.GroupType(typeStr)
		if !groupType.IsValid() {
			return entities.GroupFilter{}, httperrors.NewBadRequest(fmt.Sprintf("invalid type: %s", typeStr))
		}
		filter.Type = groupType
	}

	if activeStr := c.QueryParam("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err != nil {
			return entities.GroupFilter{}, httperrors.NewBadRequest("active must be true or false")
		}
		filter.Active = &active
	}

	if searchStr := c.QueryParam("q"); searchStr != "" {
		filter.Search = searchStr
	}

	if deletedStr := c.QueryParam("deleted"); deletedStr != "" {
		deleted, err := strconv.ParseBool(deletedStr)
		if err != nil {
			return entities.GroupFilter{}, httperrors.NewBadRequest("deleted must be true or false")
		}
		filter.Deleted = deleted
	}

	return filter, nil
}

func (h *Handler) CreateGroup(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := userIDFromContext(c)
	if err != nil {
		return err
	}

	req, err := makeGroupRequestFromContext(c, userID, h.log)
	if err != nil {
		return httperrors.NewBadRequest(err.Error())
	}

	groupID, providerCode, err := h.svc.CreateGroup(ctx, req)
	if err != nil {
		return httperrors.NewInternal(err)
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
		return httperrors.NewBadRequest("invalid request body")
	}
	userID, err := userIDFromContext(c)
	if err != nil {
		return err
	}
	req.OwnerID = userID

	if raw := c.FormValue("miembros"); raw != "" && len(req.Members) == 0 {
		if err := json.Unmarshal([]byte(raw), &req.Members); err != nil {
			return httperrors.NewBadRequest("miembros: formato inválido")
		}
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	group := updateGroupEntityFromRequest(req)

	if logo, err := utils.GetFileFrom(c, entities.GroupFileTypeLogo); err == nil && logo != nil {
		group.Logo = logo
	}
	if projectFile, err := utils.GetFileFrom(c, entities.GroupFileTypeProject); err == nil && projectFile != nil {
		group.Project = projectFile
	}

	for i := range group.Members {
		if doc, err := utils.GetFileFrom(c, fmt.Sprintf("documento_miembro_%d", i)); err == nil && doc != nil {
			group.Members[i].Document = doc
		}
	}

	if err = h.svc.UpdateGroup(ctx, req.ID, group); err != nil {
		return mapGroupError(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteGroup(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteGroupRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	userID, err := userIDFromContext(c)
	if err != nil {
		return err
	}
	req.OwnerID = userID
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	if err = h.svc.DeleteGroup(ctx, req.ID, req.OwnerID); err != nil {
		return mapGroupError(err)
	}

	return c.JSON(http.StatusNoContent, nil)
}

func makeGroupRequestFromContext(c echo.Context, userID string, log *zap.Logger) (entities.ExtensionGroup, error) {
	form, err := c.MultipartForm()
	var faculties []string
	var types []string
	if err == nil {
		faculties = form.Value["facultad"]
		types = form.Value["tipo"]
	}
	if len(faculties) == 0 {
		if f := c.FormValue("facultad"); f != "" {
			faculties = append(faculties, f)
		}
	}
	if len(types) == 0 {
		if t := c.FormValue("tipo"); t != "" {
			types = append(types, t)
		}
	}

	req := CreateGroupRequest{
		Name:                c.FormValue("nombre"),
		Description:         c.FormValue("descripcion"),
		Type:                types,
		Foundation:          c.FormValue("fundacion"),
		IsMultidisciplinary: c.FormValue("es_multidisciplinario") == "true",
		Faculty:             faculties,
		Objective:           c.FormValue("objetivo"),
		Location:            c.FormValue("ubicacion"),
		Email:               c.FormValue("correo"),
		Phone:               c.FormValue("telefono"),
	}

	if raw := c.FormValue("miembros"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &req.Members); err != nil {
			log.Error("error parsing members", zap.Error(err))
			return entities.ExtensionGroup{}, errors.New("miembros: formato inválido")
		}
	}

	if err := c.Validate(req); err != nil {
		log.Error("error validating request", zap.Error(err))
		return entities.ExtensionGroup{}, err
	}

	logo, err := utils.GetFileFrom(c, entities.GroupFileTypeLogo)
	if err != nil {
		log.Error("error getting logo", zap.Error(err))
		return entities.ExtensionGroup{}, errors.New("logo is required")
	}

	projectFile, err := utils.GetFileFrom(c, entities.GroupFileTypeProject)
	if err != nil {
		log.Error("error getting project file", zap.Error(err))
		return entities.ExtensionGroup{}, errors.New("project file is required")
	}

	group := createGroupEntityFromRequest(req, userID, req.Faculty)
	group.Logo = logo
	group.Project = projectFile

	for i := range group.Members {
		doc, err := utils.GetFileFrom(c, fmt.Sprintf("documento_miembro_%d", i))
		if err != nil {
			log.Error("error getting member document", zap.Int("member_index", i), zap.Error(err))
			return entities.ExtensionGroup{}, fmt.Errorf("documento del miembro #%d es requerido", i+1)
		}
		group.Members[i].Document = doc
	}

	return group, nil
}
