package courses

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error)
	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error)

	// Unit of Work: Atomic operations
	CreateCourseWithRequest(ctx context.Context, course entities.Course) (courseID int64, requestID int64, err error)

	// Individual operations (for flexibility)
	CreateCourse(ctx context.Context, course entities.Course) (int64, error)
	UpdateCourse(ctx context.Context, courseID string, user entities.Course) error
	DeleteCourse(ctx context.Context, courseID string) error

	// Course Requests
	CreateCourseRequest(ctx context.Context, request entities.CourseRequest) (int64, error)
}

type CourseService struct {
	repo Repository
	log  *zap.Logger
}

func NewCoursesService(repository Repository, logger *zap.Logger) *CourseService {
	return &CourseService{
		repo: repository,
		log:  logger,
	}
}

func (s *CourseService) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	course, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return entities.Course{}, err
	}
	return course, nil
}

func (s *CourseService) GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error) {
	period, err := s.repo.GetLatestCoursePeriod(ctx, courseID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	return period, nil
}

func (s *CourseService) GetCourses(
	ctx context.Context,
	pageScope entities.PageScope,
) ([]entities.Course, entities.PageScope, error) {
	courses, page, err := s.repo.GetCourses(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return courses, page, nil
}

func (s *CourseService) CreateCourse(ctx context.Context, userID string, course entities.Course) (int64, error) {
	provider, err := s.repo.GetProviderByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get provider by user ID", zap.Error(err), zap.String("user_id", userID))
		return -1, err
	}

	course.Owner.ID = provider.ID
	courseID, requestID, err := s.repo.CreateCourseWithRequest(ctx, course)
	if err != nil {
		s.log.Error("failed to create course with request", zap.Error(err))
		return -1, err
	}

	s.log.Info("course and course request created successfully",
		zap.Int64("course_id", courseID),
		zap.Int64("request_id", requestID))

	return courseID, nil
}

func (s *CourseService) UpdateCourse(
	ctx context.Context,
	courseID string,
	course entities.Course,
) error {
	if err := s.repo.UpdateCourse(ctx, courseID, course); err != nil {
		return err
	}
	return nil
}

func (s *CourseService) DeleteCourse(ctx context.Context, courseID string) error {
	if err := s.repo.DeleteCourse(ctx, courseID); err != nil {
		return err
	}
	return nil
}
