package course_requests

import (
	"context"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	ApproveCourseRequest(ctx context.Context, cr entities.CourseRequest, courseType entities.CourseType) error
	RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error
	RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error
	GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error)
}

type CourseRequestEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeCourseRequestEndpointsHandler(svc Service, log *zap.Logger) *CourseRequestEndpointsHandler {
	return &CourseRequestEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *CourseRequestEndpointsHandler) RegisterCourseRequestAdminEndpoints(g *echo.Group) {
	cr := g.Group("/course-requests")
	cr.GET("", h.GetCourseRequestsByFaculty)
	cr.GET("/:id", h.GetCourseRequestByID)
	cr.POST("/:id/approve", h.ApproveCourseRequest)
	cr.POST("/:id/reject", h.RejectCourseRequest)
	cr.POST("/:id/redirect", h.RedirectCourseRequest)
}

func (h *CourseRequestEndpointsHandler) ApproveCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	var req ApproveCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode request body", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	request := entities.CourseRequest{
		ID:       reqID,
		Reviewer: entities.User{ID: userID},
		Comments: req.Comments,
	}
	courseType := entities.FromStringCourseType(req.CourseType)

	if err := h.svc.ApproveCourseRequest(ctx, request, courseType); err != nil {
		h.log.Error("failed to approve course request", zap.Error(err), zap.String("id", reqID))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *CourseRequestEndpointsHandler) RejectCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	var req RejectCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode request body", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := h.svc.RejectCourseRequest(ctx, reqID, userID, req.Comments); err != nil {
		h.log.Error("failed to reject course request", zap.Error(err), zap.String("id", reqID))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *CourseRequestEndpointsHandler) RedirectCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		h.log.Error("no user ID found in context")
		return echo.NewHTTPError(http.StatusUnauthorized, "user missing from context")
	}

	var req RedirectCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode request body", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(req); err != nil {
		h.log.Error("invalid request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate faculty using entities.FromString
	faculty, err := entities.FromString(req.Faculty)
	if err != nil {
		h.log.Error("invalid faculty", zap.Error(err), zap.String("faculty", req.Faculty))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid faculty")
	}

	if err := h.svc.RedirectCourseRequest(ctx, reqID, userID, faculty, req.Reason); err != nil {
		h.log.Error("failed to redirect course request", zap.Error(err), zap.String("id", reqID))
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *CourseRequestEndpointsHandler) GetCourseRequestsByFaculty(c echo.Context) error {
	ctx := c.Request().Context()
	faculty, err := entities.FromString(c.QueryParam("faculty"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.ErrBadFaculty)
	}

	var pageScope entities.PageScope
	if err := pageScope.GetPageFromVars(c.QueryParam("page")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number")
	}

	if err := pageScope.GetPerPageFromVars(c.QueryParam("pageSize")); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page size")
	}

	requests, resultScope, err := h.svc.GetCourseRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		h.log.Error("failed to get course requests", zap.Error(err), zap.String("faculty", string(faculty)))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get course requests")
	}

	response := GetCourseRequestsResponse{
		CourseRequests: make([]GetCourseRequestResponse, 0, len(requests)),
		Pages:          resultScope,
	}

	for _, req := range requests {
		response.CourseRequests = append(response.CourseRequests, GetCourseRequestResponse{
			ID:          req.ID,
			Status:      req.Status,
			Comments:    req.Comments,
			ReviewedAt:  req.ReviewedAt,
			CreatedAt:   req.CreatedAt,
			UpdatedAtAt: req.UpdatedAtAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *CourseRequestEndpointsHandler) GetCourseRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request ID is required")
	}

	req, err := h.svc.GetCourseRequestByID(ctx, reqID)
	if err != nil {
		h.log.Error("failed to get course request", zap.Error(err), zap.String("id", reqID))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get course request")
	}

	response := GetCourseRequestResponse{
		ID:          reqID,
		Status:      req.Status,
		Comments:    req.Comments,
		ReviewedAt:  req.ReviewedAt,
		CreatedAt:   req.CreatedAt,
		UpdatedAtAt: req.UpdatedAtAt,
	}

	// Map course if present
	if req.Course != nil {
		response.Course = &CourseInfo{
			ID:          req.Course.ID,
			Name:        req.Course.Name,
			Description: req.Course.Description,
			Objectives:  req.Course.Objectives,
			Duration:    req.Course.Duration,
			Content:     req.Course.Content,
			Type:        req.Course.Type.String(),
			Faculty:     req.Course.Faculty.String(),
			Cost:        req.Course.Cost,
			Location:    req.Course.Location,
			CreatedAt:   req.Course.CreatedAt,
			UpdatedAt:   req.Course.UpdatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}
