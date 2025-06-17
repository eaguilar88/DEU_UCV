package providers

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

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
	GetFilesByOwner(ctx context.Context, ownerID string) (entities.GroupedFiles, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, file io.Reader, objectKey string, metadata map[string]string) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var provider entities.Provider
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider by ID", zap.Error(err), zap.String("provider_id", providerID))
		return entities.Provider{}, fmt.Errorf("failed to get provider by ID: %w", err)
	}

	files, err := s.getFilesForProvider(ctx, providerID)
	if err != nil {
		return entities.Provider{}, err
	}

	provider.Files = files
	return provider, nil
}

func (s *ProvidersService) GetProviderByCode(ctx context.Context, code string) (entities.Provider, error) {
	provider, err := s.repo.GetProviderByCode(ctx, code)
	if err != nil {
		s.logger.Error("failed to get provider by code", zap.Error(err))
		return entities.Provider{}, err
	}
	files, err := s.getFilesForProvider(ctx, provider.ID)
	if err != nil {
		s.logger.Error("failed to get files for provider", zap.Error(err), zap.String("provider_id", provider.ID))
		return entities.Provider{}, err
	}
	provider.Files = files
	return provider, nil
}

func (s *ProvidersService) GetProviders(ctx context.Context, pageScope entities.PageScope) ([]entities.Provider, entities.PageScope, error) {
	providers, pageScope, err := s.repo.GetProviders(ctx, pageScope)
	if err != nil {
		s.logger.Error("failed to get providers", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	var wg sync.WaitGroup
	for i := range providers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			files, err := s.getFilesForProvider(ctx, providers[idx].ID)
			if err != nil {
				s.logger.Error("failed to get files for provider", zap.Error(err), zap.String("provider_id", providers[idx].ID))
				return
			}
			providers[idx].Files = files
		}(i)
	}
	wg.Wait()

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

	commonMetadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_code":    provider.Code,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}
	providerIDInt, err := strconv.ParseInt(providerID, 10, 64)
	if err != nil {
		s.logger.Error("invalid providerID", zap.Error(err))
		return err
	}
	files, err := prepareFilesSlice(provider, providerIDInt, commonMetadata)
	if err != nil {
		s.logger.Error("failed to prepare files map", zap.Error(err))
		return err
	}
	if err = s.uploadAndSave(ctx, files); err != nil {
		s.logger.Error("failed to upload files file", zap.Error(err))
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

func (s *ProvidersService) uploadAndSave(ctx context.Context, files []*entities.File) error {
	for _, file := range files {
		file.MetaData["file_purpose"] = string(file.Purpose)
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

func prepareFilesSlice(provider *entities.Provider, providerID int64, metadata map[string]string) ([]*entities.File, error) {
	filesArr := []*entities.File{
		makeFileEntityFromFilePointer(provider.Files.CI, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.Files.RIF, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.Files.ISLR, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
	}

	for _, resume := range provider.Files.Resumes {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(resume, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	for _, other := range provider.Files.Others {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(other, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	return filesArr, nil
}

func makeFileEntityFromFilePointer(file *entities.File, providerID int64, uploadedBy string, ownerType entities.OwnerType, metadata map[string]string) *entities.File {
	file.OwnerID = fmt.Sprintf("%d", providerID)
	file.OwnerType = ownerType
	file.Key = fmt.Sprintf("files/providers/%d/%s", providerID, file.Name)
	file.Public = false
	file.MetaData = metadata
	file.UploadedBy = uploadedBy
	file.CreatedAt = time.Now().Format(time.RFC3339)
	return file
}

// getFilesForProvider retrieves and populates file URLs for a provider.
func (s *ProvidersService) getFilesForProvider(ctx context.Context, providerID string) (entities.ProviderFiles, error) {
	files, err := s.repo.GetFilesByOwner(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get files by owner", zap.Error(err), zap.String("provider_id", providerID))
		return entities.ProviderFiles{}, fmt.Errorf("failed to get files by owner: %w", err)
	}

	var wgFiles sync.WaitGroup
	for _, file := range files.GetAllFiles() {
		if file == nil {
			continue
		}
		wgFiles.Add(1)
		go func(f *entities.File) {
			defer wgFiles.Done()
			url, err := s.storage.GetFileURL(ctx, f.Key)
			if err != nil {
				s.logger.Error("failed to get file URL", zap.Error(err), zap.String("provider_id", providerID), zap.String("file_key", f.Key))
				return
			}
			f.URL = url
		}(file)
	}
	wgFiles.Wait()

	return entities.ProviderFiles{
		CI:      files.GetSingleFile(entities.ProviderFileTypeCI),
		RIF:     files.GetSingleFile(entities.ProviderFileTypeRIF),
		ISLR:    files.GetSingleFile(entities.ProviderFileTypeISLR),
		Resumes: files.GetMultipleFiles(entities.ProviderFileTypeResume),
		Others:  files.GetMultipleFiles(entities.ProviderFileTypeOther),
	}, nil
}
