package providers

import (
	"context"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/storage"
	"go.uber.org/zap"
)

// TODO: Implement service.go logic

type Repository interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, code string) (entities.Provider, error)
	GetProviders(
		ctx context.Context,
		pageScope entities.PageScope,
	) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
}

type ProvidersService struct {
	repo    Repository
	storage storage.StorageClient
	logger  *zap.Logger
}

func NewProvidersService(
	repo Repository,
	storage storage.StorageClient,
	logger *zap.Logger,
) Service {
	return &ProvidersService{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

func (s *ProvidersService) GetProvider(
	ctx context.Context,
	providerID string,
) (entities.Provider, error) {
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider", zap.Error(err))
		return entities.Provider{}, err
	}
	return provider, nil
}

func (s *ProvidersService) GetProviderByCode(
	ctx context.Context,
	code string,
) (entities.Provider, error) {
	provider, err := s.repo.GetProviderByCode(ctx, code)
	if err != nil {
		s.logger.Error("failed to get provider by code", zap.Error(err))
		return entities.Provider{}, err
	}
	return provider, nil
}

func (s *ProvidersService) GetProviders(
	ctx context.Context,
	pageScope entities.PageScope,
) ([]entities.Provider, entities.PageScope, error) {
	providers, pageScope, err := s.repo.GetProviders(ctx, pageScope)
	if err != nil {
		s.logger.Error("failed to get providers", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return providers, pageScope, nil
}

func (s *ProvidersService) CreateProvider(
	ctx context.Context,
	provider *entities.Provider,
) (int64, error) {
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
	metadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_code":    provider.Code,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}

	if createdProviderID > 0 {
		// Upload files to storage
		ci, err := provider.CI.Open()
		if err != nil {
			s.logger.Error("failed to open CI file", zap.Error(err))
			return -1, err
		}
		if err := s.storage.UploadFile(ctx, ci, fmt.Sprintf("providers/%d/files/%s", createdProviderID, provider.CI.Filename), metadata); err != nil {
			s.logger.Error("failed to upload CI file", zap.Error(err))
			return -1, err
		}
		// if err := s.storage.UploadFile(ctx, provider.RIF, createdProviderID); err != nil {
		// 	s.logger.Error("failed to upload RIF file", zap.Error(err))
		// 	return -1, err
		// }
		// for _, resume := range provider.Resumes {
		// 	if err := s.storage.UploadFile(ctx, resume, createdProviderID); err != nil {
		// 		s.logger.Error("failed to upload resume file", zap.Error(err))
		// 		return -1, err
		// 	}
		// }
		// for _, other := range provider.Others {
		// 	if err := s.storage.UploadFile(ctx, other, createdProviderID); err != nil {
		// 		s.logger.Error("failed to upload other file", zap.Error(err))
		// 		return -1, err
		// 	}
		// }
	}
	return createdProviderID, nil
}

func (s *ProvidersService) UpdateProvider(
	ctx context.Context,
	providerID string,
	provider *entities.Provider,
) error {
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
