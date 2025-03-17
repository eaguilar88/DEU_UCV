package courses

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
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	CreateCourse(ctx context.Context, course entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID int, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID int) error
}

type CourseEndpointsHandler struct {
	svc Service
	log *zap.Logger
}

func MakeCourseEndpointsHandler(svc Service, log *zap.Logger) CourseEndpointsHandler {
	return CourseEndpointsHandler{
		svc: svc,
		log: log,
	}
}

func (h *CourseEndpointsHandler) GetCourse(c echo.Context) error {
	ctx := c.Request().Context()
	req := GetCourseRequest{ID: c.Param("id")}
	course, err := h.svc.GetCourse(ctx, req.ID)
	if err != nil {
		h.log.Error(fmt.Sprintf("error getting course with ID: %s", req.ID), zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, EntitiesCourseToGetCourseResponse(course))
}

func (h *CourseEndpointsHandler) GetCourses(c echo.Context) error {
	ctx := c.Request().Context()
	scope := entities.PageScope{}

	scope.GetPageFromVars(c.QueryParam("page"))
	scope.GetPerPageFromVars(c.QueryParam("per_page"))
	req := GetCoursesRequest{
		PageScope: scope,
	}
	courses, pages, err := h.svc.GetCourses(ctx, req.PageScope)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, GetCoursesResponse{
		Courses: EntitiesCoursesToGetCoursesResponse(courses),
		Pages:   pages,
	})
}

func (h *CourseEndpointsHandler) CreateCourse(c echo.Context) error {
	ctx := c.Request().Context()
	var req CreateCourseRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrBadRequest
	}

	courseID, err := h.svc.CreateCourse(ctx, createCourseRequestToEntitiesCourse(req))
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusCreated, CreateCoursesResponse{
		ID: fmt.Sprintf("%d", courseID),
	})
}

func (h *CourseEndpointsHandler) UpdateCourse(c echo.Context) error {
	ctx := c.Request().Context()
	var req UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("invalid id", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	updatedCourse := updateCourseRequestToEntitiesCourse(req, intID)
	err = h.svc.UpdateCourse(ctx, intID, updatedCourse)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}

func (h *CourseEndpointsHandler) DeleteCourse(c echo.Context) error {
	ctx := c.Request().Context()
	var req DeleteCourseRequest
	if err := c.Bind(&req); err != nil {
		h.log.Error("could not decode", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	intID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("invalid id", zap.Error(err), zap.Any("request", req))
		return echo.ErrBadRequest
	}

	err = h.svc.DeleteCourse(ctx, intID)
	if err != nil {
		h.log.Error("could not decode", zap.Error(err))
		return echo.ErrUnprocessableEntity
	}

	return c.JSON(http.StatusAccepted, nil)
}
