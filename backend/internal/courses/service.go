package courses

import (
	"context"
	"fmt"
	"io"
	"time"

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

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

// StorageClient defines the file storage operations required by the courses service.
type StorageClient interface {
	UploadFile(ctx context.Context, file []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

type service struct {
	repo    Repository
	storage StorageClient
	log     *zap.Logger
}

func NewService(repository Repository, storage StorageClient, logger *zap.Logger) Service {
	return &service{
		repo:    repository,
		storage: storage,
		log:     logger,
	}
}

func (s *service) GetCourse(ctx context.Context, courseID string) (entities.Course, error) {
	course, err := s.repo.GetCourse(ctx, courseID)
	if err != nil {
		return entities.Course{}, err
	}

	files, err := s.repo.GetFilesByOwner(ctx, courseID, entities.OwnerTypeCourse)
	if err != nil {
		s.log.Error("failed to get course files", zap.Error(err), zap.String("course_id", courseID))
		return entities.Course{}, err
	}
	if cover := files.GetSingleFile(entities.CourseFileTypeCover); cover != nil {
		url, err := s.storage.GetFileURL(ctx, cover.Key)
		if err != nil {
			s.log.Error("failed to get cover URL", zap.Error(err), zap.String("course_id", courseID))
			return entities.Course{}, err
		}
		cover.URL = url
		course.Cover = cover
	}

	return course, nil
}

func (s *service) GetLatestCoursePeriod(ctx context.Context, courseID string) (entities.CoursePeriod, error) {
	period, err := s.repo.GetLatestCoursePeriod(ctx, courseID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	return period, nil
}

func (s *service) GetCourses(ctx context.Context, pageScope entities.PageScope) ([]entities.Course, entities.PageScope, error) {
	courses, page, err := s.repo.GetCourses(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	for i := range courses {
		files, err := s.repo.GetFilesByOwner(ctx, courses[i].ID, entities.OwnerTypeCourse)
		if err != nil {
			s.log.Error("failed to get course files", zap.Error(err), zap.String("course_id", courses[i].ID))
			continue
		}
		if cover := files.GetSingleFile(entities.CourseFileTypeCover); cover != nil {
			url, err := s.storage.GetFileURL(ctx, cover.Key)
			if err != nil {
				s.log.Error("failed to get cover URL", zap.Error(err), zap.String("course_id", courses[i].ID))
				continue
			}
			cover.URL = url
			courses[i].Cover = cover
		}
	}

	return courses, page, nil
}

func (s *service) CreateCourse(ctx context.Context, userID string, course entities.Course) (int64, error) {
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

	if course.Cover != nil {
		courseIDStr := fmt.Sprintf("%d", courseID)
		course.Cover.OwnerID = courseIDStr
		course.Cover.OwnerType = entities.OwnerTypeCourse
		course.Cover.Key = fmt.Sprintf("files/courses/%d/%s", courseID, course.Cover.Name)
		course.Cover.Purpose = entities.CourseFileTypeCover
		course.Cover.Public = false
		course.Cover.UploadedBy = userID
		course.Cover.CreatedAt = time.Now().Format(time.RFC3339)

		if err := s.storage.UploadFile(ctx, []*entities.File{course.Cover}); err != nil {
			s.log.Error("failed to upload course cover", zap.Error(err), zap.String("course_id", courseIDStr))
			return -1, err
		}

		if err := s.repo.SaveFilesToDB(ctx, []*entities.File{course.Cover}); err != nil {
			s.log.Error("failed to save course cover metadata", zap.Error(err), zap.String("course_id", courseIDStr))
			return -1, err
		}

		s.log.Info("course cover uploaded successfully", zap.String("course_id", courseIDStr))
	}

	return courseID, nil
}

func (s *service) UpdateCourse(
	ctx context.Context,
	courseID string,
	course entities.Course,
) error {
	return s.repo.UpdateCourse(ctx, courseID, course)
}

func (s *service) DeleteCourse(ctx context.Context, courseID string) error {
	return s.repo.DeleteCourse(ctx, courseID)
}
