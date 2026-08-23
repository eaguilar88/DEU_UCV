package course_periods

import (
	"context"
	"errors"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Repository interface {
	GetCourse(ctx context.Context, courseID string) (entities.Course, error)
	GetActiveCoursePeriodByCourseID(ctx context.Context, courseID string) (entities.CoursePeriod, error)
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

type service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(repository Repository, logger *zap.Logger) Service {
	return &service{
		repo: repository,
		log:  logger,
	}
}

func (s *service) GetCoursePeriod(ctx context.Context, periodID string) (entities.CoursePeriod, error) {
	period, err := s.repo.GetCoursePeriodByID(ctx, periodID)
	if err != nil {
		return entities.CoursePeriod{}, err
	}

	var announcements []entities.Announcement
	announcements, err = s.repo.GetAnnouncementsByCoursePeriodID(ctx, periodID)
	if err != nil {
		s.log.Warn("could not get announcements for course period", zap.String("periodID", periodID), zap.Error(err))
	} else {
		period.Announcements = announcements
	}

	return period, nil
}

func (s *service) GetCoursePeriods(ctx context.Context, courseID string, pageScope entities.PageScope) ([]entities.CoursePeriod, entities.PageScope, error) {
	periods, page, err := s.repo.GetCoursePeriods(ctx, courseID, pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := range periods {
		i := i
		g.Go(func() error {
			announcements, err := s.repo.GetAnnouncementsByCoursePeriodID(ctx, periods[i].ID)
			if err != nil {
				return err
			}
			periods[i].Announcements = announcements
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, entities.PageScope{}, err
	}

	return periods, page, nil
}

func (s *service) CreateCoursePeriod(ctx context.Context, period entities.CoursePeriod) (int64, error) {
	// security.CustomValidator silently drops regular field-validation errors, so the
	// "capacidad" validate tag on the request DTO is not actually enforced at the HTTP layer.
	if period.Capacity <= 0 {
		return -1, ErrInvalidCapacity
	}

	if _, err := s.repo.GetCourse(ctx, period.Course.ID); err != nil {
		return -1, err
	}

	if _, err := s.repo.GetActiveCoursePeriodByCourseID(ctx, period.Course.ID); err != nil {
		if !errors.Is(err, ErrCoursePeriodNotFound) {
			return -1, err
		}
	} else {
		return -1, ErrCoursePeriodAlreadyOpen
	}

	id, err := s.repo.CreateCoursePeriod(ctx, period)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *service) UpdateCoursePeriod(ctx context.Context, periodID string, period entities.CoursePeriod) error {
	return s.repo.UpdateCoursePeriod(ctx, periodID, period)
}

func (s *service) DeleteCoursePeriod(
	ctx context.Context,
	periodID, userID string,
) error {
	return s.repo.DeleteCoursePeriod(ctx, periodID)
}

// Announcement methods
func (s *service) GetAnnouncement(ctx context.Context, announcementID string) (entities.Announcement, error) {
	announcement, err := s.repo.GetAnnouncementByID(ctx, announcementID)
	if err != nil {
		return entities.Announcement{}, err
	}
	return announcement, nil
}

func (s *service) CreateAnnouncement(ctx context.Context, periodID string, announcement entities.Announcement) (int64, error) {
	id, err := s.repo.CreateAnnouncement(ctx, periodID, announcement)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (s *service) UpdateAnnouncement(ctx context.Context, announcementID string, announcement entities.Announcement) error {
	return s.repo.UpdateAnnouncement(ctx, announcementID, announcement)
}

func (s *service) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	return s.repo.DeleteAnnouncement(ctx, announcementID)
}
