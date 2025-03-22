package course_periods

// TODO: Implement service.go logic
import (
	"context"
	"strconv"

	"github.com/eaguilar88/deu/pkg/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCoursePeriodByID(ctx context.Context, periodID int) (entities.CoursePeriod, error)
	GetCoursePeriods(ctx context.Context, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
	CreateCoursePeriod(ctx context.Context, coursePeriod entities.CoursePeriod) (int64, error)
	UpdateCoursePeriod(ctx context.Context, periodID int, coursePeriod entities.CoursePeriod) error
	DeleteCoursePeriod(ctx context.Context, periodID int) error
	GetUsersByCoursePeriodID(ctx context.Context, periodID int) ([]entities.User, error)
}

type CoursePeriodService struct {
	repo Repository
	log  *zap.Logger
}

func NewCoursePeriodService(repository Repository, logger *zap.Logger) *CoursePeriodService {
	return &CoursePeriodService{
		repo: repository,
		log:  logger,
	}
}

func (s *CoursePeriodService) GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
	intID, err := strconv.Atoi(periodID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	period, err := s.repo.GetCoursePeriodByID(ctx, intID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}
	participantsChan := make(chan []entities.User)
	errorChan := make(chan error)

	go func() {
		users, err := s.repo.GetUsersByCoursePeriodID(ctx, intID)
		if err != nil {
			errorChan <- err
			return
		}
		participantsChan <- users
	}()

	select {
	case participants := <-participantsChan:
		period.Participants = participants
	case err := <-errorChan:
		return entities.CoursePeriod{}, err
	}
	return period, nil
}

func (s *CoursePeriodService) GetCoursePeriods(ctx context.Context, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	periods, page, err := s.repo.GetCoursePeriods(ctx, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}
	return periods, page, nil
}

func (s *CoursePeriodService) CreateCoursePeriod(ctx context.Context, period entities.CoursePeriod) (int64, error) {
	id, err := s.repo.CreateCoursePeriod(ctx, period)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *CoursePeriodService) UpdateCoursePeriod(ctx context.Context, periodID int, period entities.CoursePeriod) error {
	if err := s.repo.UpdateCoursePeriod(ctx, periodID, period); err != nil {
		return err
	}
	return nil
}

func (s *CoursePeriodService) DeleteCoursePeriod(ctx context.Context, periodID int) error {
	if err := s.repo.DeleteCoursePeriod(ctx, periodID); err != nil {
		return err
	}
	return nil
}
