package notifications

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

const dueDateLayout = "02/01/2006"

type Repository interface {
	CountStaleCourseRequestsByFaculty(ctx context.Context, createdBefore time.Time) ([]entities.FacultyPendingCount, error)
	CountStaleGroupRequestsByFaculty(ctx context.Context, createdBefore time.Time) ([]entities.FacultyPendingCount, error)
	GetFacultyCoordinatorEmails(ctx context.Context, faculty entities.Faculty) ([]string, error)
	GetDEUAdminEmails(ctx context.Context) ([]string, error)
	GetGroupsDueForRenewal(ctx context.Context, dueBefore time.Time) ([]entities.GroupRenewalDue, error)
	DeactivateExpiredGroups(ctx context.Context, now time.Time) ([]entities.GroupRenewalDue, error)
	GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error)
	WasNotifiedSince(ctx context.Context, kind, target, recipient string, since time.Time) (bool, error)
	LogNotification(ctx context.Context, n entities.NotificationLog) error
}

// MailClient defines the email sending operations required by the notifications service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

// Config sets when reminders are due.
type Config struct {
	// PendingRequestAge is how old a request under review must be before its coordinators are reminded.
	PendingRequestAge time.Duration
	// RenewalWindow is how long before a group's renewal date its contact starts being reminded.
	RenewalWindow time.Duration
	// ResendInterval is how long to wait before repeating a reminder to the same recipient.
	ResendInterval time.Duration
}

type Service struct {
	repo        Repository
	emailClient MailClient
	config      Config
	now         func() time.Time
	logger      *zap.Logger
}

func NewService(repo Repository, emailClient MailClient, config Config, logger *zap.Logger) *Service {
	return &Service{
		repo:        repo,
		emailClient: emailClient,
		config:      config,
		now:         time.Now,
		logger:      logger,
	}
}

// Run deactivates the groups whose renewal expired and sends every reminder that is due.
func (s *Service) Run(ctx context.Context) error {
	return errors.Join(
		s.DeactivateExpiredGroups(ctx),
		s.SendCoordinatorReminders(ctx),
		s.SendGroupRenewalReminders(ctx),
	)
}

// DeactivateExpiredGroups deactivates the active groups whose yearly renewal date passed without
// an approved renewal. Approving a renewal reactivates the group.
func (s *Service) DeactivateExpiredGroups(ctx context.Context) error {
	expired, err := s.repo.DeactivateExpiredGroups(ctx, s.now())
	if err != nil {
		s.logger.Error("failed to deactivate expired groups", zap.Error(err))
		return fmt.Errorf("deactivating expired groups: %w", err)
	}
	for _, g := range expired {
		s.logger.Info("group deactivated: renewal expired",
			zap.String("group_id", g.GroupID),
			zap.String("group_name", g.GroupName),
			zap.Time("renewal_due_at", g.RenewalDueAt))
	}
	return nil
}

// SendCoordinatorReminders emails each faculty's coordinators a summary of the course and group
// requests that have been under review for at least PendingRequestAge. DEU requests also go to
// the DEU admins.
func (s *Service) SendCoordinatorReminders(ctx context.Context) error {
	createdBefore := s.now().Add(-s.config.PendingRequestAge)

	courses, err := s.repo.CountStaleCourseRequestsByFaculty(ctx, createdBefore)
	if err != nil {
		s.logger.Error("failed to count pending course requests", zap.Error(err))
		return fmt.Errorf("counting pending course requests: %w", err)
	}
	groups, err := s.repo.CountStaleGroupRequestsByFaculty(ctx, createdBefore)
	if err != nil {
		s.logger.Error("failed to count pending group requests", zap.Error(err))
		return fmt.Errorf("counting pending group requests: %w", err)
	}

	pending := make(map[entities.Faculty]*email.CoordinatorPendingReminderData)
	summary := func(f entities.Faculty) *email.CoordinatorPendingReminderData {
		if pending[f] == nil {
			pending[f] = &email.CoordinatorPendingReminderData{
				Faculty: f.String(),
				MinDays: int(s.config.PendingRequestAge.Hours() / 24),
			}
		}
		return pending[f]
	}
	for _, c := range courses {
		summary(c.Faculty).CourseRequests += c.Count
	}
	for _, g := range groups {
		summary(g.Faculty).GroupRequests += g.Count
	}

	var errs []error
	for faculty, data := range pending {
		recipients, err := s.coordinatorEmails(ctx, faculty)
		if err != nil {
			s.logger.Error("failed to get coordinators", zap.Error(err), zap.String("faculty", faculty.String()))
			errs = append(errs, err)
			continue
		}
		if len(recipients) == 0 {
			s.logger.Warn("no coordinators to remind of pending requests", zap.String("faculty", faculty.String()))
			continue
		}
		for _, to := range recipients {
			if err := s.remind(ctx, entities.NotificationKindCoordinatorPendingRequests, faculty.String(), to, email.TemplateCoordinatorPendingReminder, *data); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// SendGroupRenewalReminders emails the contact of each approved group whose renewal is due within
// RenewalWindow, or already overdue.
func (s *Service) SendGroupRenewalReminders(ctx context.Context) error {
	now := s.now()
	groups, err := s.repo.GetGroupsDueForRenewal(ctx, now.Add(s.config.RenewalWindow))
	if err != nil {
		s.logger.Error("failed to get groups due for renewal", zap.Error(err))
		return fmt.Errorf("getting groups due for renewal: %w", err)
	}

	var errs []error
	for _, g := range groups {
		to, err := s.groupContactEmail(ctx, g.GroupID)
		if err != nil {
			s.logger.Warn("group has no contact email for its renewal reminder", zap.Error(err), zap.String("group_id", g.GroupID))
			continue
		}
		data := email.GroupRenewalReminderData{
			GroupName: g.GroupName,
			DueDate:   g.RenewalDueAt.Format(dueDateLayout),
			Overdue:   g.RenewalDueAt.Before(now),
		}
		if err := s.remind(ctx, entities.NotificationKindGroupRenewal, g.GroupID, to, email.TemplateGroupRenewalReminder, data); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// remind sends a reminder unless the same one was sent to recipient within ResendInterval, and
// records it once sent.
func (s *Service) remind(ctx context.Context, kind, target, recipient string, tmpl email.Template, data any) error {
	logger := s.logger.With(zap.String("kind", kind), zap.String("target", target), zap.String("recipient", recipient))

	sent, err := s.repo.WasNotifiedSince(ctx, kind, target, recipient, s.now().Add(-s.config.ResendInterval))
	if err != nil {
		logger.Error("failed to check notifications log", zap.Error(err))
		return fmt.Errorf("checking notifications log: %w", err)
	}
	if sent {
		return nil
	}

	if err := s.emailClient.SendTemplate(ctx, recipient, tmpl, data); err != nil {
		logger.Error("failed to send reminder", zap.Error(err))
		return fmt.Errorf("sending %s reminder: %w", kind, err)
	}

	if err := s.repo.LogNotification(ctx, entities.NotificationLog{Kind: kind, Target: target, Recipient: recipient}); err != nil {
		logger.Error("failed to log sent reminder", zap.Error(err))
		return fmt.Errorf("logging %s reminder: %w", kind, err)
	}
	logger.Info("reminder sent")
	return nil
}

func (s *Service) coordinatorEmails(ctx context.Context, faculty entities.Faculty) ([]string, error) {
	recipients, err := s.repo.GetFacultyCoordinatorEmails(ctx, faculty)
	if err != nil {
		return nil, err
	}
	if faculty != entities.FacultyDEU {
		return recipients, nil
	}

	admins, err := s.repo.GetDEUAdminEmails(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range admins {
		if !slices.Contains(recipients, a) {
			recipients = append(recipients, a)
		}
	}
	return recipients, nil
}

// groupContactEmail returns the email saved in deu.contacts when the group was registered.
func (s *Service) groupContactEmail(ctx context.Context, groupID string) (string, error) {
	contacts, err := s.repo.GetContactsByOwner(ctx, groupID, entities.OwnerTypeExtensionGroup)
	if err != nil {
		return "", err
	}
	for _, c := range contacts {
		if c.Type == entities.ContactTypeEmail && c.Value != "" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("group %s has no email contact", groupID)
}
