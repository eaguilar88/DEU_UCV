package courses

import (
	"context"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCourse(ctx context.Context, courseID int) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	CreateCourse(ctx context.Context, course entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID int, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID int) error
}

type CourseService struct {
	repo Repository
	log  *zap.Logger
}

func NewCoursesService(repository Repository, logger *zap.Logger) Service {
	return &CourseService{
		repo: repository,
		log:  logger,
	}
}

func (s *CourseService) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	intID, err := strconv.Atoi(courseID)
	if err != nil {
		return entities.Course{}, err
	}
	course, err := s.repo.GetCourse(ctx, intID)
	if err != nil {
		return entities.Course{}, err
	}
	return course, nil
}

func (s *CourseService) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	courses, page, err := s.repo.GetCourses(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return courses, page, nil
}

func (s *CourseService) CreateCourse(ctx context.Context, course entities.Course) (int64, error) {
	id, err := s.repo.CreateCourse(ctx, course)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *CourseService) UpdateCourse(ctx context.Context, courseID int, course entities.Course) error {
	if err := s.repo.UpdateCourse(ctx, courseID, course); err != nil {
		return err
	}
	return nil
}

func (s *CourseService) DeleteCourse(ctx context.Context, courseID int) error {
	if err := s.repo.DeleteCourse(ctx, courseID); err != nil {
		return err
	}
	return nil
}
