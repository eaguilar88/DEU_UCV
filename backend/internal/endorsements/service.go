package endorsements

import (
	"context"
	"mime/multipart"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetEndorsement(ctx context.Context, endorsementID string) (entities.CourseRequest, error)
	GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error)
	CreateEndorsement(ctx context.Context, endorsement entities.CourseRequest) (int64, error)
	UpdateEndorsement(ctx context.Context, endorsementID string, endorsement entities.CourseRequest) error
	DeleteEndorsement(ctx context.Context, endorsementID string) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file multipart.File, objectKey string, metadata map[string]string) error
	DownloadFile(ctx context.Context, objectKey string, destinationPath string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
}

type EndorsementService struct {
	repo    Repository
	storage StorageClient
	log     *zap.Logger
}

func NewEndorsementsService(repository Repository, s3 StorageClient, logger *zap.Logger) *EndorsementService {
	return &EndorsementService{
		repo:    repository,
		storage: s3,
		log:     logger,
	}
}

func (s *EndorsementService) GetEndorsement(ctx context.Context, endorsementID string) (entities.CourseRequest, error) {
	user, err := s.repo.GetEndorsement(ctx, endorsementID)
	if err != nil {
		return entities.CourseRequest{}, err
	}
	return user, nil
}

func (s *EndorsementService) GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.CourseRequest, entities.PageScope, error) {
	users, page, err := s.repo.GetEndorsements(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return users, page, nil
}

func (s *EndorsementService) CreateEndorsement(ctx context.Context, endorsement entities.CourseRequest) (int64, error) {
	id, err := s.repo.CreateEndorsement(ctx, endorsement)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *EndorsementService) UpdateEndorsement(ctx context.Context, endorsement entities.CourseRequest) error {
	if err := s.repo.UpdateEndorsement(ctx, endorsement.ID, endorsement); err != nil {
		return err
	}
	return nil
}

func (s *EndorsementService) DeleteEndorsement(ctx context.Context, endorsementID string) error {
	if err := s.repo.DeleteEndorsement(ctx, endorsementID); err != nil {
		return err
	}
	return nil
}
