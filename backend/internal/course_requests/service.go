package course_requests

import (
	"context"
	"mime/multipart"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCourseRequest(ctx context.Context, courseRequestID string) (entities.CourseRequest, error)
	GetCourseRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	CreateCourseRequest(ctx context.Context, courseRequest entities.CourseRequest) (int64, error)
	UpdateCourseRequest(ctx context.Context, courseRequestID string, courseRequest entities.CourseRequest) error
	DeleteCourseRequest(ctx context.Context, courseRequestID string) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file multipart.File, objectKey string, metadata map[string]string) error
	DownloadFile(ctx context.Context, objectKey string, destinationPath string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
}

type CourseRequestService struct {
	repo    Repository
	storage StorageClient
	log     *zap.Logger
}

func NewCourseRequestService(repository Repository, s3 StorageClient, logger *zap.Logger) *CourseRequestService {
	return &CourseRequestService{
		repo:    repository,
		storage: s3,
		log:     logger,
	}
}

func (s *CourseRequestService) GetCourseRequest(ctx context.Context, courseRequestID string) (entities.CourseRequest, error) {
	user, err := s.repo.GetCourseRequest(ctx, courseRequestID)
	if err != nil {
		return entities.CourseRequest{}, err
	}
	return user, nil
}

func (s *CourseRequestService) GetCourseRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	users, page, err := s.repo.GetCourseRequests(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return users, page, nil
}

func (s *CourseRequestService) CreateCourseRequest(ctx context.Context, courseRequest entities.CourseRequest) (int64, error) {
	id, err := s.repo.CreateCourseRequest(ctx, courseRequest)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *CourseRequestService) UpdateCourseRequest(ctx context.Context, courseRequest entities.CourseRequest) error {
	if err := s.repo.UpdateCourseRequest(ctx, courseRequest.ID, courseRequest); err != nil {
		return err
	}
	return nil
}

func (s *CourseRequestService) DeleteCourseRequest(ctx context.Context, courseRequestID string) error {
	if err := s.repo.DeleteCourseRequest(ctx, courseRequestID); err != nil {
		return err
	}
	return nil
}
