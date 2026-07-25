package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetActivityByID(ctx context.Context, id string) (entities.Activity, error)
	GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error)
	CreateActivity(ctx context.Context, activity entities.Activity) (int64, error)
	UpdateActivity(ctx context.Context, activity entities.Activity) error
	DeleteActivity(ctx context.Context, id string) error
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, files []*entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
}

type service struct {
	repo    Repository
	storage StorageClient
	logger  *zap.Logger
}

func NewService(repo Repository, storage StorageClient, logger *zap.Logger) Service {
	return &service{repo: repo, storage: storage, logger: logger}
}

func (s *service) GetActivity(ctx context.Context, id string) (entities.Activity, error) {
	activity, err := s.repo.GetActivityByID(ctx, id)
	if err != nil {
		return entities.Activity{}, err
	}

	files, err := s.repo.GetFilesByOwner(ctx, id, entities.OwnerTypeActivity)
	if err != nil {
		s.logger.Error("failed to get activity files", zap.Error(err), zap.String("activity_id", id))
		return entities.Activity{}, err
	}

	if cover := files.GetSingleFile(entities.ActivityFileTypeCoverImage); cover != nil {
		url, err := s.storage.GetFileURL(ctx, cover.Key)
		if err != nil {
			s.logger.Error("failed to get cover image URL", zap.Error(err), zap.String("activity_id", id))
			return entities.Activity{}, err
		}
		cover.URL = url
		activity.CoverImage = cover
	}

	return activity, nil
}

func (s *service) GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
	return s.repo.GetActivities(ctx, filter, pageScope)
}

func (s *service) CreateActivity(ctx context.Context, activity entities.Activity) (int64, error) {
	activityID, err := s.repo.CreateActivity(ctx, activity)
	if err != nil {
		s.logger.Error("failed to create activity", zap.Error(err))
		return -1, err
	}

	if activity.CoverImage == nil {
		return activityID, nil
	}

	activityIDStr := fmt.Sprintf("%d", activityID)
	activity.CoverImage.OwnerID = activityIDStr
	activity.CoverImage.OwnerType = entities.OwnerTypeActivity
	activity.CoverImage.Purpose = entities.ActivityFileTypeCoverImage
	activity.CoverImage.Key = fmt.Sprintf("files/activities/%d/%s", activityID, activity.CoverImage.Name)
	activity.CoverImage.Public = false
	activity.CoverImage.CreatedAt = time.Now().Format(time.RFC3339)

	if err := s.storage.UploadFile(ctx, []*entities.File{activity.CoverImage}); err != nil {
		s.logger.Error("failed to upload activity cover image", zap.Error(err), zap.String("activity_id", activityIDStr))
		return -1, err
	}

	if err := s.repo.SaveFilesToDB(ctx, []*entities.File{activity.CoverImage}); err != nil {
		s.logger.Error("failed to save activity cover image to DB", zap.Error(err), zap.String("activity_id", activityIDStr))
		return -1, err
	}

	return activityID, nil
}

func (s *service) UpdateActivity(ctx context.Context, id string, activity entities.Activity) error {
	activity.ID = id
	return s.repo.UpdateActivity(ctx, activity)
}

func (s *service) DeleteActivity(ctx context.Context, id string) error {
	return s.repo.DeleteActivity(ctx, id)
}
