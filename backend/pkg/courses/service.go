package courses

import (
	"context"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
)

type Repository interface {
	GetCourse(ctx context.Context, userID int) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	CreateCourse(ctx context.Context, user entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, userID int, user entities.Course) error
	DeleteCourse(ctx context.Context, userID int) error
}

type CourseService struct {
	repo Repository
	log  log.Logger
}

func NewCoursesService(repository Repository, logger log.Logger) Service {
	return &CourseService{
		repo: repository,
		log:  logger,
	}
}

func (s *CourseService) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	panic("unimplemented")
}

func (s *CourseService) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	panic("unimplemented")
}

func (s *CourseService) CreateCourse(ctx context.Context, user entities.Course) (int64, error) {
	panic("unimplemented")
}

func (s *CourseService) UpdateCourse(ctx context.Context, courseID int, user entities.Course) error {
	panic("unimplemented")
}

func (s *CourseService) DeleteCourse(ctx context.Context, courseID int) error {
	panic("unimplemented")
}
