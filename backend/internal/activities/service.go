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

	for _, f := range files.GetAllFiles() {
		url, err := s.storage.GetFileURL(ctx, f.Key)
		if err != nil {
			s.logger.Error("failed to get file URL", zap.Error(err), zap.String("key", f.Key))
			continue
		}
		f.URL = url
	}
	activity.Files = files

	return activity, nil
}

func (s *service) GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, error) {
	return s.repo.GetActivities(ctx, filter, pageScope)
}

func (s *service) CreateActivity(ctx context.Context, activity entities.Activity, files []*entities.File) (int64, error) {
	activityID, err := s.repo.CreateActivity(ctx, activity)
	if err != nil {
		s.logger.Error("failed to create activity", zap.Error(err))
		return -1, err
	}

	if len(files) == 0 {
		return activityID, nil
	}

	activityIDStr := fmt.Sprintf("%d", activityID)
	now := time.Now().Format(time.RFC3339)
	for _, f := range files {
		f.OwnerID = activityIDStr
		f.OwnerType = entities.OwnerTypeActivity
		f.Purpose = entities.ActivityFileTypeReport
		f.Key = fmt.Sprintf("files/activities/%d/%s", activityID, f.Name)
		f.Public = false
		f.CreatedAt = now
	}

	if err := s.storage.UploadFile(ctx, files); err != nil {
		s.logger.Error("failed to upload activity files", zap.Error(err), zap.String("activity_id", activityIDStr))
		return -1, err
	}

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.logger.Error("failed to save activity files to DB", zap.Error(err), zap.String("activity_id", activityIDStr))
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
