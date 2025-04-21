package groups

import (
	"context"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

// TODO: Implement service.go logic
type Repository interface {
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, error)
	UpdateGroup(ctx context.Context, group entities.ExtensionGroup) error
	DeleteGroup(ctx context.Context, groupID string) error
}

type GroupService struct {
	repo Repository
	log  *zap.Logger
}

func NewGroupsService(repository Repository, logger *zap.Logger) Service {
	return &GroupService{
		repo: repository,
		log:  logger,
	}
}

func (s *GroupService) GetGroup(
	ctx context.Context,
	groupID string,
) (entities.ExtensionGroup, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}
	return group, nil
}

func (s *GroupService) GetGroups(
	ctx context.Context,
	pageScope entities.PageScope,
) ([]entities.ExtensionGroup, entities.PageScope, error) {
	groups, page, err := s.repo.GetGroups(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return groups, page, nil
}

func (s *GroupService) CreateGroup(
	ctx context.Context,
	group entities.ExtensionGroup,
	userID string,
) (int64, error) {
	id, err := s.repo.CreateGroup(ctx, group)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *GroupService) UpdateGroup(
	ctx context.Context,
	groupID string,
	group entities.ExtensionGroup,
) error {
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
