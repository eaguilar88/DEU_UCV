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

const maxFeaturedActivities = 4

type service struct {
	repo    Repository
	storage StorageClient
	logger  *zap.Logger
}

func NewService(repo Repository, storage StorageClient, logger *zap.Logger) Service {
	return &service{repo: repo, storage: storage, logger: logger}
}

// buildActivityFile fills in the ownership/storage metadata for a file about to be
// uploaded for an activity. It mutates and returns file, or returns nil if file is nil,
// so call sites can conditionally append without a separate nil check.
func buildActivityFile(file *entities.File, ownerID, purpose, keyPrefix, uploadedBy, createdAt string, public bool) *entities.File {
	if file == nil {
		return nil
	}
	file.OwnerID = ownerID
	file.OwnerType = entities.OwnerTypeActivity
	file.Purpose = purpose
	file.Key = keyPrefix + file.Name
	file.Public = public
	file.UploadedBy = uploadedBy
	file.CreatedAt = createdAt
	return file
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

	activityIDStr := fmt.Sprintf("%d", activityID)
	now := time.Now().Format(time.RFC3339)

	var filesToUpload []*entities.File
	if f := buildActivityFile(activity.CoverImage, activityIDStr, entities.ActivityFileTypeCoverImage,
		fmt.Sprintf("files/activities/%s/", activityIDStr), activity.CreatedBy, now, true); f != nil {
		filesToUpload = append(filesToUpload, f)
	}
	if f := buildActivityFile(activity.ParticipantList, activityIDStr, entities.ActivityFileTypeListParticipants,
		fmt.Sprintf("files/activities/%s/participants_", activityIDStr), activity.CreatedBy, now, false); f != nil {
		filesToUpload = append(filesToUpload, f)
	}

	if len(filesToUpload) == 0 {
		return activityID, nil
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

	now := time.Now().Format(time.RFC3339)

	var filesToUpload []*entities.File
	if f := buildActivityFile(activity.CoverImage, id, entities.ActivityFileTypeCoverImage,
		fmt.Sprintf("files/activities/%s/", id), activity.CreatedBy, now, true); f != nil {
		filesToUpload = append(filesToUpload, f)
	}
	if f := buildActivityFile(activity.ParticipantList, id, entities.ActivityFileTypeListParticipants,
		fmt.Sprintf("files/activities/%s/participants_", id), activity.CreatedBy, now, false); f != nil {
		filesToUpload = append(filesToUpload, f)
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

		if count >= maxFeaturedActivities {
			return ErrMaxFeaturedLimitReached
		}
	}
	return s.repo.UpdateFeatureStatus(ctx, id, featured)
}

func (s *service) DeleteActivity(ctx context.Context, id string) error {
	return s.repo.DeleteActivity(ctx, id)
}
