package providers

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

// TODO: Implement service.go logic

type Repository interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, code string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string) ([]entities.File, error)
	SaveFilesToDB(ctx context.Context, file []entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file multipart.File, objectKey string, metadata map[string]string) error
	DownloadFile(ctx context.Context, objectKey string, destinationPath string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
}

type ProvidersService struct {
	repo    Repository
	storage StorageClient
	logger  *zap.Logger
}

func NewProvidersService(repo Repository, storage StorageClient, logger *zap.Logger) *ProvidersService {
	return &ProvidersService{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

func (s *ProvidersService) GetProvider(ctx context.Context, providerID string) (entities.Provider, error) {
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider", zap.Error(err))
		return entities.Provider{}, err
	}
	return provider, nil
}

func (s *ProvidersService) GetProviderByCode(ctx context.Context, code string) (entities.Provider, error) {
	provider, err := s.repo.GetProviderByCode(ctx, code)
	if err != nil {
		s.logger.Error("failed to get provider by code", zap.Error(err))
		return entities.Provider{}, err
	}
	return provider, nil
}

func (s *ProvidersService) GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
	providers, pageScope, err := s.repo.GetProviders(ctx, pageScope)
	if err != nil {
		s.logger.Error("failed to get providers", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return providers, pageScope, nil
}

func (s *ProvidersService) CreateProvider(ctx context.Context, provider *entities.Provider) (int64, error) {
	code, err := entities.GenerateProviderCode(provider.Type)
	if err != nil {
		s.logger.Error("failed to generate provider code", zap.Error(err))
		return -1, err
	}
	provider.Code = code
	createdProviderID, err := s.repo.CreateProvider(ctx, *provider)
	if err != nil {
		s.logger.Error("failed to create provider", zap.Error(err))
		return -1, err
	}

	commonMetadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_code":    provider.Code,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}
	files, err := prepareFilesSlice(provider, createdProviderID, commonMetadata)
	if err != nil {
		s.logger.Error("failed to prepare files map", zap.Error(err))
		return -1, err
	}
	if err = s.uploadAndSave(ctx, files); err != nil {
		s.logger.Error("failed to upload files file", zap.Error(err))
		return -1, err
	}

	return createdProviderID, nil
}

func (s *ProvidersService) UpdateProvider(ctx context.Context, providerID string, provider *entities.Provider) error {
	err := s.repo.UpdateProvider(ctx, providerID, *provider)
	if err != nil {
		s.logger.Error("failed to update provider", zap.Error(err))
		return err
	}
	return nil
}

func (s *ProvidersService) DeleteProvider(ctx context.Context, providerID string) error {
	err := s.repo.DeleteProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to delete provider", zap.Error(err))
		return err
	}
	return nil
}

func (s *ProvidersService) uploadAndSave(ctx context.Context, files []entities.File) error {
	for _, file := range files {
		if err := s.storage.UploadFile(ctx, file.Body, file.Key, file.MetaData); err != nil {
			s.logger.Error("failed to upload file", zap.Error(err), zap.String("file_key", file.Key))
			return err
		}
	}

	err := s.repo.SaveFilesToDB(ctx, files)
	if err != nil {
		s.logger.Error("failed to save file metadata to database", zap.Error(err))
		return err
	}
	return nil
}

func prepareFilesSlice(provider *entities.Provider, providerID int64, metadata map[string]string) ([]entities.File, error) {
	filesArr := []entities.File{
		makeFileEntityFromFilePointer(provider.CI, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.RIF, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.ISLR, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
	}

	for _, resume := range provider.Resumes {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(resume, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	for _, other := range provider.Others {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(other, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	return filesArr, nil
}

func makeFileEntityFromFilePointer(fileHeader *multipart.FileHeader, providerID int64, uploadedBy string, ownerType entities.OwnerType, metadata map[string]string) entities.File {
	body, err := fileHeader.Open()
	if err != nil {
		return entities.File{}
	}
	return entities.File{
		OwnerID:    fmt.Sprintf("%d", providerID),
		OwnerType:  ownerType,
		Key:        fmt.Sprintf("files/providers/%d/files/%s", providerID, fileHeader.Filename),
		Body:       body,
		Public:     false,
		MetaData:   metadata,
		UploadedBy: uploadedBy,
	}
}
