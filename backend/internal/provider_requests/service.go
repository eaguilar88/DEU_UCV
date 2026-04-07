package provider_requests

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

var (
	ErrRequestIsProcessed = errors.New("esta solicitud ya ha sido procesada por un administrador")
)

type Repository interface {
	GetProviderRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.ProviderRequest, entities.PageScope, error)
	GetProviderRequestByID(ctx context.Context, id string) (entities.ProviderRequest, error)
	ApproveProviderRequest(ctx context.Context, id, reviewerID string, providerID int64, code string) error
	RejectProviderRequest(ctx context.Context, id, reviewerID, comments string) error
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
}

type MailClient interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type service struct {
	repo        Repository
	emailClient MailClient
	logger      *zap.Logger
}

func NewService(repo Repository, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repo,
		emailClient: emailClient,
		logger:      logger,
	}
}

func (s *service) ApproveProviderRequest(ctx context.Context, id, reviewerID string) error {
	existing, err := s.repo.GetProviderRequestByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get provider request", zap.Error(err))
		return err
	}
	if existing.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	providerIDStr := strconv.FormatInt(existing.ProviderID, 10)
	provider, err := s.repo.GetProvider(ctx, providerIDStr)
	if err != nil {
		s.logger.Error("failed to get provider", zap.Error(err), zap.String("provider_id", providerIDStr))
		return err
	}

	code, err := entities.GenerateProviderCode(provider.Type)
	if err != nil {
		s.logger.Error("failed to generate provider code", zap.Error(err))
		return err
	}

	if err := s.repo.ApproveProviderRequest(ctx, id, reviewerID, existing.ProviderID, code); err != nil {
		s.logger.Error("failed to approve provider request", zap.Error(err))
		return err
	}

	if err := s.emailClient.Send(ctx, provider.User.Email, "Solicitud de registro aprobada", fmt.Sprintf("Tu solicitud de registro como proveedor ha sido aprobada. Tu código de proveedor es: %s", code)); err != nil {
		s.logger.Error("failed to send approval email", zap.Error(err))
		return fmt.Errorf("error sending approval email: %w", err)
	}

	return nil
}

func (s *service) RejectProviderRequest(ctx context.Context, id, reviewerID, comments string) error {
	existing, err := s.repo.GetProviderRequestByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get provider request", zap.Error(err))
		return err
	}
	if existing.Status != entities.RequestStatus_UNDER_REVIEW {
		return ErrRequestIsProcessed
	}

	return s.repo.RejectProviderRequest(ctx, id, reviewerID, comments)
}

func (s *service) GetProviderRequests(ctx context.Context, pageScope entities.PageScope) ([]entities.ProviderRequest, entities.PageScope, error) {
	requests, ps, err := s.repo.GetProviderRequests(ctx, pageScope)
	if err != nil {
		s.logger.Error("failed to get provider requests", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return requests, ps, nil
}

func (s *service) GetProviderRequestByID(ctx context.Context, id string) (entities.ProviderRequest, error) {
	return s.repo.GetProviderRequestByID(ctx, id)
}
