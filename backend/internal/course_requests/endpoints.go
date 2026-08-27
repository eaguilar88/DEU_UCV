package course_requests

import (
	"context"
	"errors"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/facultyscope"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	ApproveCourseRequest(ctx context.Context, cr entities.CourseRequest, courseType entities.CourseType) error
	RejectCourseRequest(ctx context.Context, reqID, reviewerID, comments string) error
	RedirectCourseRequest(ctx context.Context, reqID, reviewerID string, faculty entities.Faculty, reason string) error
	GetCourseRequestsByFaculty(ctx context.Context, faculty entities.Faculty, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	GetCourseRequestByID(ctx context.Context, reqID string) (entities.CourseRequest, error)
	GetMyCourseRequests(ctx context.Context, userID string, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
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

func (h *Handler) RegisterCourseRequestAdminEndpoints(g *echo.Group) {
	cr := g.Group("/course-requests")
	cr.GET("", h.GetCourseRequestsByFaculty)
	cr.GET("/:id", h.GetCourseRequestByID)
	cr.POST("/:id/approve", h.ApproveCourseRequest)
	cr.POST("/:id/reject", h.RejectCourseRequest)
	cr.POST("/:id/redirect", h.RedirectCourseRequest)
}

func (h *Handler) RegisterCourseRequestEndpoints(g *echo.Group) {
	g.GET("/course-requests", h.GetMyCourseRequests)
}

func (h *Handler) ApproveCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req ApproveCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	request := entities.CourseRequest{
		ID:       reqID,
		Reviewer: entities.User{ID: userID},
		Comments: req.Comments,
	}
	courseType := entities.FromStringCourseType(req.CourseType)

	if err := h.svc.ApproveCourseRequest(ctx, request, courseType); err != nil {
		if errors.Is(err, ErrCourseRequestNotFound) {
			return httperrors.NewNotFound("course request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RejectCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req RejectCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}

	if err := h.svc.RejectCourseRequest(ctx, reqID, userID, req.Comments); err != nil {
		if errors.Is(err, ErrCourseRequestNotFound) {
			return httperrors.NewNotFound("course request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) RedirectCourseRequest(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var req RedirectCourseRequestRequest
	if err := c.Bind(&req); err != nil {
		return httperrors.NewBadRequest("invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return httperrors.NewBadRequest("validation failed")
	}

	// Validate faculty using entities.FromString
	faculty, err := entities.FromString(req.Faculty)
	if err != nil {
		return httperrors.NewBadRequest("invalid faculty")
	}

	if err := h.svc.RedirectCourseRequest(ctx, reqID, userID, faculty, req.Reason); err != nil {
		if errors.Is(err, ErrCourseRequestNotFound) {
			return httperrors.NewNotFound("course request not found")
		}
		return httperrors.NewInternal(err)
	}

	return c.NoContent(http.StatusAccepted)
}

func (h *Handler) GetCourseRequestsByFaculty(c echo.Context) error {
	ctx := c.Request().Context()
	faculty, err := facultyscope.Resolve(c, c.QueryParam("faculty"))
	if err != nil {
		return httperrors.NewBadRequest("invalid faculty")
	}

	var pageScope entities.PageScope
	if err := pageScope.GetPageFromVars(c.QueryParam("page")); err != nil {
		return httperrors.NewBadRequest("invalid page number")
	}

	if err := pageScope.GetPerPageFromVars(c.QueryParam("pageSize")); err != nil {
		return httperrors.NewBadRequest("invalid page size")
	}

	requests, resultScope, err := h.svc.GetCourseRequestsByFaculty(ctx, faculty, pageScope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, courseRequestsToResponse(requests, resultScope))
}

func (h *Handler) GetMyCourseRequests(c echo.Context) error {
	ctx := c.Request().Context()
	userID, ok := c.Get("userID").(string)
	if !ok {
		return httperrors.NewUnauthorized("authentication required")
	}

	var pageScope entities.PageScope
	if err := pageScope.GetPageFromVars(c.QueryParam("page")); err != nil {
		return httperrors.NewBadRequest("invalid page number")
	}

	if err := pageScope.GetPerPageFromVars(c.QueryParam("pageSize")); err != nil {
		return httperrors.NewBadRequest("invalid page size")
	}

	requests, resultScope, err := h.svc.GetMyCourseRequests(ctx, userID, pageScope)
	if err != nil {
		return httperrors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, courseRequestsToResponse(requests, resultScope))
}

// courseRequestsToResponse converts a slice of CourseRequest entities to the shared
// list-response shape, used by both the admin (faculty-scoped) and the plain
// (submitter-scoped) course-requests listing endpoints so their formats stay identical.
func courseRequestsToResponse(requests []entities.CourseRequest, pageScope entities.PageScope) GetCourseRequestsResponse {
	response := GetCourseRequestsResponse{
		CourseRequests: make([]GetCourseRequestResponse, 0, len(requests)),
		Pages:          pageScope,
	}

	for _, req := range requests {
		response.CourseRequests = append(response.CourseRequests, GetCourseRequestResponse{
			ID:         req.ID,
			Status:     req.Status,
			Comments:   req.Comments,
			ReviewedAt: req.ReviewedAt,
			CreatedAt:  req.CreatedAt,
			UpdatedAt:  req.UpdatedAt,
		})
	}

	return response
}

func (h *Handler) GetCourseRequestByID(c echo.Context) error {
	ctx := c.Request().Context()
	reqID := c.Param("id")
	if reqID == "" {
		return httperrors.NewBadRequest("request ID is required")
	}

	req, err := h.svc.GetCourseRequestByID(ctx, reqID)
	if err != nil {
		if errors.Is(err, ErrCourseRequestNotFound) {
			return httperrors.NewNotFound("course request not found")
		}
		return httperrors.NewInternal(err)
	}

	response := GetCourseRequestResponse{
		ID:         reqID,
		Status:     req.Status,
		Comments:   req.Comments,
		ReviewedAt: req.ReviewedAt,
		CreatedAt:  req.CreatedAt,
		UpdatedAt:  req.UpdatedAt,
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
