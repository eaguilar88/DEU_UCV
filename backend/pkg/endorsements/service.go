package endorsements

import (
	"context"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/go-kit/log"
)

type Repository interface {
	GetEndorsement(ctx context.Context, endorsementID int) (entities.Endorsements, error)
	GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsements, entities.PageScope, error)
	CreateEndorsement(ctx context.Context, endorsement entities.Endorsements) (int64, error)
	UpdateEndorsement(ctx context.Context, endorsementID int, endorsement entities.Endorsements) error
	DeleteEndorsement(ctx context.Context, endorsementID int) error
}

type EndorsementService struct {
	repo Repository
	log  log.Logger
}

func NewEndorsementsService(repository Repository, logger log.Logger) *EndorsementService {
	return &EndorsementService{
		repo: repository,
		log:  logger,
	}
}

func (s *EndorsementService) GetEndorsement(ctx context.Context, endorsementID string) (entities.Endorsements, error) {
	intID, err := strconv.Atoi(endorsementID)
	if err != nil {
		return entities.Endorsements{}, err
	}
	user, err := s.repo.GetEndorsement(ctx, intID)
	if err != nil {
		return entities.Endorsements{}, err
	}
	return user, nil
}
func (s *EndorsementService) GetEndorsements(ctx context.Context, pageScope entities.PageScope) ([]entities.Endorsements, entities.PageScope, error) {
	users, page, err := s.repo.GetEndorsements(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return users, page, nil
}

func (s *EndorsementService) CreateEndorsement(ctx context.Context, endorsement entities.Endorsements) (int64, error) {
	id, err := s.repo.CreateEndorsement(ctx, endorsement)
	if err != nil {
		return -1, err
	}
	return id, nil
}
func (s *EndorsementService) UpdateEndorsement(ctx context.Context, endorsementID int, endorsement entities.Endorsements) error {
	if err := s.repo.UpdateEndorsement(ctx, endorsementID, endorsement); err != nil {
		return err
	}
	return nil
}
func (s *EndorsementService) DeleteEndorsement(ctx context.Context, endorsementID int) error {
	if err := s.repo.DeleteEndorsement(ctx, endorsementID); err != nil {
		return err
	}
	return nil
}
