package groups

import (
	"context"
	"fmt"
	"io"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)

	// Unit of Work: Atomic operations
	CreateGroupWithRequests(ctx context.Context, group entities.ExtensionGroup, requests []entities.GroupRequest) (int64, error)

	// Individual operations (for flexibility)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, error)
	UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID string) error

	// Requests
	CreateGroupRequest(ctx context.Context, req entities.GroupRequest) (int64, error)

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file io.Reader, objectKey string, metadata map[string]string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

type service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(repository Repository, logger *zap.Logger) Service {
	return &service{
		repo: repository,
		log:  logger,
	}
}

func (s *service) GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return group, nil
}

func (s *service) GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	groups, page, err := s.repo.GetGroups(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return groups, page, nil
}

func (s *service) CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, string, error) {
	userID := group.Owner.ID
	providerCode, err := entities.GenerateProviderCode(entities.GroupProviderType)
	if err != nil {
		s.log.Error("failed to generate provider code",
			zap.Error(err),
			zap.String("action", "generate_code"),
		)
		return -1, "", fmt.Errorf("failed to generate provider code: %w", err)
	}

	// Build approval requests based on group type
	base := entities.GroupRequest{
		Status:   entities.RequestStatus_UNDER_REVIEW,
		Comments: "New group creation request",
	}
	var requests []entities.GroupRequest
	if group.Type == entities.GroupType(entities.MultidisciplinaryGroupType) {
		deuReq := base
		deuReq.Faculty = entities.FacultyDEU
		requests = []entities.GroupRequest{deuReq}
	} else {
		facultyReq := base
		facultyReq.Faculty = group.Faculty
		deuReq := base
		deuReq.Faculty = entities.FacultyDEU
		requests = []entities.GroupRequest{facultyReq, deuReq}
	}

	// Create group and requests in a single transaction
	groupID, err := s.repo.CreateGroupWithRequests(ctx, group, requests)
	if err != nil {
		s.log.Error("failed to create group with requests",
			zap.Error(err),
			zap.String("action", "create_group_with_requests"),
			zap.String("group_name", group.Name),
		)
		return -1, "", fmt.Errorf("failed to create group with requests: %w", err)
	}

	// Prepare files metadata
	commonMetadata := map[string]string{
		"group_owner":   group.ID,
		"group_id":      fmt.Sprintf("%d", groupID),
		"group_name":    group.Name,
		"provider_code": providerCode,
	}

	files := []*entities.File{
		makeFileEntityFromFilePointer(group.Files.Logo, groupID, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
		// makeFileEntityFromFilePointer(group.Files.FinancingPlan, groupID, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
		// makeFileEntityFromFilePointer(group.Files.GroupProject, groupID, userID, entities.OwnerTypeExtensionGroup, commonMetadata),
	}

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.log.Error("failed to save files",
			zap.Error(err),
			zap.String("group_id", fmt.Sprintf("%d", groupID)),
			zap.String("action", "save_files"),
		)
		return -1, "", fmt.Errorf("failed to save files: %w", err)
	}

	s.log.Info("group, requests, and files created successfully",
		zap.Int64("group_id", groupID),
		zap.Int("file_count", len(files)))

	return groupID, providerCode, nil
}

func (s *service) UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error {
	return s.repo.UpdateGroup(ctx, group)
}

func (s *service) DeleteGroup(ctx context.Context, groupID, userID string) error {
	return s.repo.DeleteGroup(ctx, groupID)
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
