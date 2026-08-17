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
	GetActivityMetrics(ctx context.Context, groupID string) (entities.ActivityMetrics, error)
	GetGroupDashboardSummary(ctx context.Context, groupID string) (entities.GroupDashboardSummary, error)
	CreateActivity(ctx context.Context, activity entities.Activity) (int64, error)
	UpdateActivity(ctx context.Context, activity entities.Activity) error
	UpdateReportCheckStatus(ctx context.Context, id string, checked bool) error
	UpdateFeatureStatus(ctx context.Context, id string, featured bool) error
	CountActivities(ctx context.Context, filter entities.ActivityFilter) (int, error)
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

	if list := files.GetSingleFile(entities.ActivityFileTypeListParticipants); list != nil {
		url, err := s.storage.GetFileURL(ctx, list.Key)
		if err == nil {
			list.URL = url
			activity.ParticipantList = list
		}
	}

	return activity, nil
}

func (s *service) GetActivities(ctx context.Context, filter entities.ActivityFilter, pageScope entities.PageScope) ([]entities.Activity, entities.PageScope, entities.ActivityMetrics, error) {
	activities, scope, err := s.repo.GetActivities(ctx, filter, pageScope)
	if err != nil {
		return nil, scope, entities.ActivityMetrics{}, err
	}

	for i := range activities {
		files, err := s.repo.GetFilesByOwner(ctx, activities[i].ID, entities.OwnerTypeActivity)
		if err != nil {
			s.logger.Error("failed to get files for activity", zap.Error(err), zap.String("activity_id", activities[i].ID))
			continue
		}

		if cover := files.GetSingleFile(entities.ActivityFileTypeCoverImage); cover != nil {
			url, err := s.storage.GetFileURL(ctx, cover.Key)
			if err == nil {
				cover.URL = url
				activities[i].CoverImage = cover
			}
		}
	}

	metrics, err := s.repo.GetActivityMetrics(ctx, filter.GroupID)
	if err != nil {
		s.logger.Error("failed to get activity metrics", zap.Error(err), zap.String("group_id", filter.GroupID))
	}

	return activities, scope, metrics, nil
}

func (s *service) GetGroupDashboardSummary(ctx context.Context, groupID string) (entities.GroupDashboardSummary, error) {
	return s.repo.GetGroupDashboardSummary(ctx, groupID)
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
	var filesToUpload []*entities.File
	activity.CoverImage.OwnerID = activityIDStr
	activity.CoverImage.OwnerType = entities.OwnerTypeActivity
	activity.CoverImage.Purpose = entities.ActivityFileTypeCoverImage
	activity.CoverImage.Key = fmt.Sprintf("files/activities/%d/%s", activityID, activity.CoverImage.Name)
	activity.CoverImage.Public = true
	activity.CoverImage.UploadedBy = activity.CreatedBy
	activity.CoverImage.CreatedAt = time.Now().Format(time.RFC3339)
	filesToUpload = append(filesToUpload, activity.CoverImage)

	if activity.ParticipantList != nil {
		activity.ParticipantList.OwnerID = activityIDStr
		activity.ParticipantList.OwnerType = entities.OwnerTypeActivity
		activity.ParticipantList.Purpose = entities.ActivityFileTypeListParticipants
		activity.ParticipantList.Key = fmt.Sprintf("files/activities/%d/participants_%s", activityID, activity.ParticipantList.Name)
		activity.ParticipantList.Public = false
		activity.ParticipantList.UploadedBy = activity.CreatedBy
		activity.ParticipantList.CreatedAt = time.Now().Format(time.RFC3339)
		filesToUpload = append(filesToUpload, activity.ParticipantList)
	}

	if err := s.storage.UploadFile(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to upload activity files", zap.Error(err), zap.String("activity_id", activityIDStr))
		return -1, err
	}

	if err := s.repo.SaveFilesToDB(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to save activity files to DB", zap.Error(err), zap.String("activity_id", activityIDStr))
		return -1, err
	}

	return activityID, nil
}

func (s *service) UpdateActivity(ctx context.Context, id string, activity entities.Activity) error {
	activity.ID = id
	if err := s.repo.UpdateActivity(ctx, activity); err != nil {
		return err
	}

	var filesToUpload []*entities.File

	if activity.CoverImage != nil {
		activity.CoverImage.OwnerID = id
		activity.CoverImage.OwnerType = entities.OwnerTypeActivity
		activity.CoverImage.Purpose = entities.ActivityFileTypeCoverImage
		activity.CoverImage.Key = fmt.Sprintf("files/activities/%s/%s", id, activity.CoverImage.Name)
		activity.CoverImage.Public = true
		activity.CoverImage.UploadedBy = activity.CreatedBy
		activity.CoverImage.CreatedAt = time.Now().Format(time.RFC3339)
		filesToUpload = append(filesToUpload, activity.CoverImage)
	}

	if activity.ParticipantList != nil {
		activity.ParticipantList.OwnerID = id
		activity.ParticipantList.OwnerType = entities.OwnerTypeActivity
		activity.ParticipantList.Purpose = entities.ActivityFileTypeListParticipants
		activity.ParticipantList.Key = fmt.Sprintf("files/activities/%s/participants_%s", id, activity.ParticipantList.Name)
		activity.ParticipantList.Public = false
		activity.ParticipantList.UploadedBy = activity.CreatedBy
		activity.ParticipantList.CreatedAt = time.Now().Format(time.RFC3339)
		filesToUpload = append(filesToUpload, activity.ParticipantList)
	}

	if len(filesToUpload) > 0 {
		if err := s.storage.UploadFile(ctx, filesToUpload); err != nil {
			s.logger.Error("failed to upload updated activity files", zap.Error(err), zap.String("activity_id", id))
			return err
		}

		if err := s.repo.SaveFilesToDB(ctx, filesToUpload); err != nil {
			s.logger.Error("failed to save updated activity files to DB", zap.Error(err), zap.String("activity_id", id))
			return err
		}
	}

	return nil
}

func (s *service) ToggleReportCheck(ctx context.Context, id string, checked bool) error {
	return s.repo.UpdateReportCheckStatus(ctx, id, checked)
}

func (s *service) ToggleFeature(ctx context.Context, id string, featured bool) error {
	act, err := s.repo.GetActivityByID(ctx, id)
	if err != nil {
		return err
	}
	if act.IsFeatured == featured {
		return nil
	}
	if featured {
		isFeatured := true
		filter := entities.ActivityFilter{
			GroupID:    act.GroupID,
			IsFeatured: &isFeatured,
		}
		count, err := s.repo.CountActivities(ctx, filter)
		if err != nil {
			return err
		}

		if count >= 4 {
			return ErrMaxFeaturedLimitReached
		}
	}
	return s.repo.UpdateFeatureStatus(ctx, id, featured)
}

func (s *service) DeleteActivity(ctx context.Context, id string) error {
	return s.repo.DeleteActivity(ctx, id)
}
