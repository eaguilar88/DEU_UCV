package courses

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/errors"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service interface {
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error)
	CreateCourse(ctx context.Context, userID string, course entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID string, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID string) error
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

func (h *Handler) GetCourse(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetCourseRequest{ID: c.Param("id")}
	course, err := h.svc.GetCourse(ctx, req.ID)
	if err != nil {
		return errors.NewInternal(err)
	}

	response := courseToResponse(course)

	// Get latest course period
	latestPeriod, err := h.svc.GetLatestCoursePeriod(ctx, req.ID)
	if err != nil {
		h.log.Warn("could not get latest course period", zap.Error(err), zap.String("course_id", req.ID))
	} else if latestPeriod.ID != "" {
		response.LatestPeriod = &LatestCoursePeriodInfo{
			ID:              latestPeriod.ID,
			StartDate:       latestPeriod.StartDate,
			EndDate:         latestPeriod.EndDate,
			InscriptionDate: latestPeriod.InscriptionDate,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetCourses(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	//nolint:errcheck
	scope.GetPageFromVars(c.QueryParam("page"))
	//nolint:errcheck
	scope.GetPerPageFromVars(c.QueryParam("per_page"))
	req := GetCoursesRequest{
		PageScope: scope,
	}
	courses, pages, err := h.svc.GetCourses(ctx, req.PageScope)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusOK, GetCoursesResponse{
		Courses: coursesToResponse(courses),
		Pages:   pages,
	})
}

func (h *Handler) CreateCourse(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := c.Get("userID").(string)
	if !ok {
		return errors.NewUnauthorized("authentication required")
	}

	course, err := toCourseEntity(c)
	if err != nil {
		return errors.NewBadRequest(err.Error())
	}

	courseID, err := h.svc.CreateCourse(ctx, userID, course)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusCreated, CreateCoursesResponse{
		ID: fmt.Sprintf("%d", courseID),
	})
}

func (h *Handler) UpdateCourse(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	updatedCourse := toCourseUpdateEntity(req, c.Param("id"))
	err := h.svc.UpdateCourse(ctx, c.Param("id"), updatedCourse)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *Handler) DeleteCourse(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCourseRequest
	if err := c.Bind(&req); err != nil {
		return errors.NewBadRequest("invalid request body")
	}

	err := h.svc.DeleteCourse(ctx, req.ID)
	if err != nil {
		return errors.NewInternal(err)
	}

	return c.JSON(http.StatusAccepted, nil)
}
