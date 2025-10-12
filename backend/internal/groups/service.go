package groups

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, error)
	UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID string) error

	// Requests
	CreateGroupRequest(ctx context.Context, req entities.GroupAuthRequest) (int64, error)

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file io.Reader, objectKey string, metadata map[string]string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

type GroupService struct {
	repo Repository
	log  *zap.Logger
}

func NewGroupsService(repository Repository, logger *zap.Logger) *GroupService {
	return &GroupService{
		repo: repository,
		log:  logger,
	}
}

func (s *GroupService) GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return group, nil
}

func (s *GroupService) GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	groups, page, err := s.repo.GetGroups(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return groups, page, nil
}

func (s *GroupService) CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, string, error) {
	userID := group.Owner.ID
	providerCode, err := entities.GenerateProviderCode(entities.GroupProviderType)
	if err != nil {
		s.log.Error("failed to generate provider code",
			zap.Error(err),
			zap.String("action", "generate_code"),
		)
		return -1, "", fmt.Errorf("failed to generate provider code: %w", err)
	}

	// Create group in database
	id, err := s.repo.CreateGroup(ctx, group)
	if err != nil {
		s.log.Error("failed to create group",
			zap.Error(err),
			zap.String("action", "create_group"),
			zap.String("group_name", group.Name),
		)
		return -1, "", fmt.Errorf("failed to create group: %w", err)
	}

	// Create a group authorization request
	groupReq := entities.GroupAuthRequest{
		GroupID:   fmt.Sprintf("%d", id),
		Faculty:   group.Faculty,
		Status:    entities.RequestStatus_UNDER_REVIEW,
		Comments:  "New group creation request",
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	// Prepare files metadata
	commonMetadata := map[string]string{
		"group_owner":   group.ID,
		"group_id":      fmt.Sprintf("%d", id),
		"group_name":    group.Name,
		"provider_code": providerCode,
	}

	files := []*entities.File{
		makeFileEntityFromFilePointer(group.Files.Logo, id, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
		makeFileEntityFromFilePointer(group.Files.FinancingPlan, id, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
		makeFileEntityFromFilePointer(group.Files.GroupProject, id, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
	}

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.log.Error("failed to save files",
			zap.Error(err),
			zap.String("group_id", fmt.Sprintf("%d", id)),
			zap.String("action", "save_files"),
		)
		return -1, "", fmt.Errorf("failed to save files: %w", err)
	}

	// Save metadata to database
	s.log.Debug("saving file metadata to database",
		zap.Int("file_count", len(files)),
		zap.String("action", "save_metadata"),
	)

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.log.Error("failed to save file metadata to database",
			zap.Error(err),
			zap.String("action", "save_metadata"),
		)
		return -1, "", fmt.Errorf("failed to save file metadata: %w", err)
	}

	s.log.Debug("successfully uploaded and saved files",
		zap.Int("file_count", len(files)),
		zap.String("action", "upload_and_save"),
	)

	grID, err := s.repo.CreateGroupRequest(ctx, groupReq)
	if err != nil {
		s.log.Error("failed to create group request",
			zap.Error(err),
			zap.String("action", "create_group_request"),
			zap.String("group_id", fmt.Sprintf("%d", id)),
		)
		return -1, "", fmt.Errorf("failed to create group request: %w", err)
	}

	return grID, providerCode, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error {
	if err := s.repo.UpdateGroup(ctx, group); err != nil {
		return err
	}
	return nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, groupID, userID string) error {
	if err := s.repo.DeleteGroup(ctx, groupID); err != nil {
		return err
	}
	return nil
}

func makeFileEntityFromFilePointer(file *entities.File, groupID int64, uploadedBy string, ownerType entities.OwnerType, metadata map[string]string) *entities.File {
	file.OwnerID = fmt.Sprintf("%d", groupID)
	file.OwnerType = ownerType
	file.Key = fmt.Sprintf("files/groups/%d/%s", groupID, file.Name)
	file.Public = false
	file.MetaData = metadata
	file.UploadedBy = uploadedBy
	return file
}
