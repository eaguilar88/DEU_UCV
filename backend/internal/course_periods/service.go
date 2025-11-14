package course_periods

import (
	"context"
	"sync"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	GetCoursePeriodByID(ctx context.Context, periodID string) (entities.CoursePeriod, error)
	GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error)
	CreateCoursePeriod(ctx context.Context, coursePeriod entities.CoursePeriod) (int64, error)
	UpdateCoursePeriod(ctx context.Context, periodID string, coursePeriod entities.CoursePeriod) error
	DeleteCoursePeriod(ctx context.Context, periodID string) error
	GetUsersByCoursePeriodID(ctx context.Context, periodID string) ([]entities.User, error)
	GetAnnouncementsByCoursePeriodID(ctx context.Context, periodID string) ([]entities.Announcement, error)
	GetAnnouncementByID(ctx context.Context, announcementID string) (entities.Announcement, error)
	CreateAnnouncement(ctx context.Context, periodID string, announcement entities.Announcement) (int64, error)
	UpdateAnnouncement(ctx context.Context, announcementID string, announcement entities.Announcement) error
	DeleteAnnouncement(ctx context.Context, announcementID string) error
}

type CoursePeriodService struct {
	repo Repository
	log  *zap.Logger
}

func NewCoursePeriodsService(repository Repository, logger *zap.Logger) *CoursePeriodService {
	return &CoursePeriodService{
		repo: repository,
		log:  logger,
	}
}

func (s *CoursePeriodService) GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
	period, err := s.repo.GetCoursePeriodByID(ctx, periodID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}

	// Get announcements in goroutine
	errCh := make(chan error)
	var announcements []entities.Announcement

	go func() {
		var err error
		announcements, err = s.repo.GetAnnouncementsByCoursePeriodID(ctx, periodID)
		errCh <- err
	}()

	// Wait for announcements
	if err := <-errCh; err != nil {
		s.log.Warn("could not get announcements for course period", zap.String("periodID", periodID), zap.Error(err))
	} else {
		period.Announcements = announcements
	}

	return period, nil
}

func (s *CoursePeriodService) GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	periods, page, err := s.repo.GetCoursePeriods(ctx, courseID, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	// Get announcements for each period using goroutines
	var wg sync.WaitGroup
	errCh := make(chan error)

	// Goroutine to collect errors
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for i := range periods {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			announcements, err := s.repo.GetAnnouncementsByCoursePeriodID(ctx, periods[i].ID)
			if err != nil {
				errCh <- err
				return
			}
			periods[i].Announcements = announcements
		}(i)
	}

	// Log any errors but don't fail the request
	for err := range errCh {
		s.log.Warn("could not get announcements for course period", zap.Error(err))
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

func (s *CoursePeriodService) UpdateCoursePeriod(ctx context.Context, periodID string, period entities.CoursePeriod) error {
	if err := s.repo.UpdateCoursePeriod(ctx, periodID, period); err != nil {
		return err
	}
	return nil
}

func (s *CoursePeriodService) DeleteCoursePeriod(
	ctx context.Context,
	periodID, userID string,
) error {
	if err := s.repo.DeleteCoursePeriod(ctx, periodID); err != nil {
		return err
	}
	return nil
}

// Announcement methods
func (s *CoursePeriodService) GetAnnouncement(ctx context.Context, announcementID string) (entities.Announcement, error) {
	announcement, err := s.repo.GetAnnouncementByID(ctx, announcementID)
	if err != nil {
		return entities.Announcement{}, err
	}
	return announcement, nil
}

func (s *CoursePeriodService) CreateAnnouncement(ctx context.Context, periodID string, announcement entities.Announcement) (int64, error) {
	id, err := s.repo.CreateAnnouncement(ctx, periodID, announcement)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *CoursePeriodService) UpdateAnnouncement(ctx context.Context, announcementID string, announcement entities.Announcement) error {
	if err := s.repo.UpdateAnnouncement(ctx, announcementID, announcement); err != nil {
		return err
	}
	return nil
}

func (s *CoursePeriodService) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	if err := s.repo.DeleteAnnouncement(ctx, announcementID); err != nil {
		return err
	}
	return nil
}
