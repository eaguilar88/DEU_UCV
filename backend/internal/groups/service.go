package groups

import (
	"context"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error)

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
	GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
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

func (s *service) GetGroup(ctx context.Context, groupID string) (entities.ExtensionGroup, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	s.enrichGroupFiles(ctx, &group)
	s.enrichGroupContacts(ctx, &group)

	return group, nil
}

func (s *service) GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error) {
	groups, page, err := s.repo.GetGroups(ctx, filter, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	for i := range groups {
		s.enrichGroupFiles(ctx, &groups[i])
		s.enrichGroupContacts(ctx, &groups[i])
	}

	return groups, page, nil
}

func (s *service) GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error) {
	groups, err := s.repo.GetRandomActiveGroups(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i := range groups {
		s.enrichGroupFiles(ctx, &groups[i])
	}

	return groups, nil
}

func (s *service) enrichGroupFiles(ctx context.Context, group *entities.ExtensionGroup) {
	if filesMap, err := s.repo.GetFilesByOwner(ctx, group.ID, entities.OwnerTypeExtensionGroup); err == nil {
		if logoList, ok := filesMap[entities.GroupFileTypeLogo]; ok && len(logoList) > 0 {
			logo := logoList[0]
			if url, err := s.storage.GetFileURL(ctx, logo.Key); err == nil {
				logo.URL = url
			}
			group.Logo = logo
		}

		if projectList, ok := filesMap[entities.GroupFileTypeProject]; ok && len(projectList) > 0 {
			project := projectList[0]
			if url, err := s.storage.GetFileURL(ctx, project.Key); err == nil {
				project.URL = url
			}
			group.Project = project
		}
	}
}

func (s *service) enrichGroupContacts(ctx context.Context, group *entities.ExtensionGroup) {
	if contacts, err := s.repo.GetContactsByOwner(ctx, group.ID, entities.OwnerTypeExtensionGroup); err == nil {
		for _, c := range contacts {
			switch c.Type {
			case entities.ContactTypeEmail:
				group.Email = c.Value
			case entities.ContactTypePhone:
				group.Phone = c.Value
			}
		}
	}
}

func hasType(types []entities.GroupType, target entities.GroupType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
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
	if group.IsMultidisciplinary || hasType(group.Type, entities.MultidisciplinaryGroupType) {
		deuReq := base
		deuReq.Faculty = entities.FacultyDEU
		requests = []entities.GroupRequest{deuReq}
	} else {
		facultyReq := base
		facultyReq.Faculty = group.Faculty[0]
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
		makeFileEntityFromFilePointer(group.Logo, groupID, userID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeLogo, commonMetadata),
		makeFileEntityFromFilePointer(group.Project, groupID, userID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeProject, commonMetadata),
	}

	validFiles := make([]*entities.File, 0, len(files))
	for _, f := range files {
		if f != nil {
			validFiles = append(validFiles, f)
		}
	}

	if len(validFiles) > 0 {
		if err := s.repo.SaveFilesToDB(ctx, validFiles); err != nil {
			s.log.Error("failed to save files",
				zap.Error(err),
				zap.String("group_id", fmt.Sprintf("%d", groupID)),
				zap.String("action", "save_files"),
			)
			return -1, "", fmt.Errorf("failed to save files: %w", err)
		}
	}

	s.log.Info("group, requests, and files created successfully",
		zap.Int64("group_id", groupID),
		zap.Int("file_count", len(validFiles)))

	return groupID, providerCode, nil
}

func (s *service) UpdateGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) error {
	err := s.repo.UpdateGroup(ctx, group)
	if err != nil {
		return err
	}

	idNum, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		return err
	}

	commonMetadata := map[string]string{
		"group_owner": group.ID,
		"group_id":    groupID,
	}

	var filesToSave []*entities.File
	if group.Logo != nil {
		filesToSave = append(filesToSave, makeFileEntityFromFilePointer(group.Logo, idNum, group.Owner.ID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeLogo, commonMetadata))
	}
	if group.Project != nil {
		filesToSave = append(filesToSave, makeFileEntityFromFilePointer(group.Project, idNum, group.Owner.ID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeProject, commonMetadata))
	}

	if len(filesToSave) > 0 {
		if err := s.repo.SaveFilesToDB(ctx, filesToSave); err != nil {
			return fmt.Errorf("failed to update group files: %w", err)
		}
	}

	return nil
}

func (s *service) DeleteGroup(ctx context.Context, groupID, userID string) error {
	return s.repo.DeleteGroup(ctx, groupID)
}

func makeFileEntityFromFilePointer(file *entities.File, groupID int64, uploadedBy string, ownerType entities.OwnerType, fileType string, metadata map[string]string) *entities.File {
	if file == nil {
		return nil
	}
	file.OwnerID = fmt.Sprintf("%d", groupID)
	file.OwnerType = ownerType

	fileMetadata := make(map[string]string)
	for k, v := range metadata {
		fileMetadata[k] = v
	}
	fileMetadata["file_type"] = fileType
	file.Key = fmt.Sprintf("files/groups/%d/%s_%s", groupID, fileType, file.Name)
	file.Public = false
	file.MetaData = fileMetadata
	file.UploadedBy = uploadedBy
	return file
}
