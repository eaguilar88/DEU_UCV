package files

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"
)

type Repository interface {
	IsPublicFile(ctx context.Context, key string) (bool, error)
}

type StorageClient interface {
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
}

type service struct {
	repo    Repository
	storage StorageClient
	logger  *zap.Logger
}

func NewService(repo Repository, storage StorageClient, logger *zap.Logger) Service {
	return &service{repo: repo, storage: storage, logger: logger}
}

// GetPublicFile returns the file only when it is flagged as public. Private files are
// delivered through pre-signed URLs, so they are reported as not found here.
func (s *service) GetPublicFile(ctx context.Context, key string) (io.ReadCloser, string, error) {
	public, err := s.repo.IsPublicFile(ctx, key)
	if err != nil {
		s.logger.Error("failed to check file visibility", zap.Error(err), zap.String("key", key))
		return nil, "", fmt.Errorf("failed to check file visibility: %w", err)
	}
	if !public {
		s.logger.Debug("refused to serve non-public file", zap.String("key", key))
		return nil, "", ErrFileNotFound
	}

	body, contentType, err := s.storage.GetObject(ctx, key)
	if err != nil {
		s.logger.Error("failed to fetch file from storage", zap.Error(err), zap.String("key", key))
		return nil, "", ErrFileNotFound
	}

	return body, contentType, nil
}
