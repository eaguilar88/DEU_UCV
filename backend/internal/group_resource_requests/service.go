package group_resource_requests

import (
	"context"
	"fmt"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type Repository interface {
	CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error)
	ApproveGroupResourceRequest(ctx context.Context, reqID string) error
	RejectGroupResourceRequest(ctx context.Context, reqID, reason string) error
	GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupResourceRequest, int, entities.PageScope, error)
	GetGroupResourceRequestsByGroupID(ctx context.Context, groupID string, status string, pageScope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error)
	GetPendingGroupResourceRequestsCountByFaculty(ctx context.Context, faculty entities.Faculty) ([]FacultyPendingCount, error)
	GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error)
	GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error)
}

// MailClient defines the email sending operations required by the group_resource_requests service.
type MailClient interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

type service struct {
	repo        Repository
	emailClient MailClient
	logger      *zap.Logger
}

func NewService(repo Repository, emailClient MailClient, logger *zap.Logger) Service {
	return &service{repo: repo, emailClient: emailClient, logger: logger}
}

func (s *service) CreateGroupResourceRequest(ctx context.Context, req entities.GroupResourceRequest) (int64, error) {
	id, err := s.repo.CreateGroupResourceRequest(ctx, req)
	if err != nil {
		s.logger.Error("failed to create group resource request", zap.Error(err))
		return 0, err
	}
	return id, nil
}

func (s *service) ApproveGroupResourceRequest(ctx context.Context, reqID string) error {
	req, err := s.repo.GetGroupResourceRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	if err := s.repo.ApproveGroupResourceRequest(ctx, reqID); err != nil {
		return err
	}

	body := fmt.Sprintf("La solicitud de recursos \"%s\" del grupo %s ha sido aprobada.", req.Type, req.GroupName)
	s.notifyGroup(ctx, req, "Solicitud de recursos aprobada", body)
	return nil
}

func (s *service) RejectGroupResourceRequest(ctx context.Context, reqID, reason string) error {
	req, err := s.repo.GetGroupResourceRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	if err := s.repo.RejectGroupResourceRequest(ctx, reqID, reason); err != nil {
		return err
	}

	body := fmt.Sprintf("La solicitud de recursos \"%s\" del grupo %s ha sido rechazada.\n\nRazón: %s", req.Type, req.GroupName, reason)
	s.notifyGroup(ctx, req, "Solicitud de recursos rechazada", body)
	return nil
}

// notifyGroup emails the group's contact address. The status change is already saved,
// so failures are logged but never returned.
func (s *service) notifyGroup(ctx context.Context, req entities.GroupResourceRequest, subject, body string) {
	logFields := []zap.Field{zap.String("group_id", req.GroupID), zap.String("request_id", req.ID)}

	contacts, err := s.repo.GetContactsByOwner(ctx, req.GroupID, entities.OwnerTypeExtensionGroup)
	if err != nil {
		s.logger.Warn("failed to get group contacts for resource request notice", append(logFields, zap.Error(err))...)
		return
	}

	to := ""
	for _, c := range contacts {
		if c.Type == entities.ContactTypeEmail && c.Value != "" {
			to = c.Value
			break
		}
	}
	if to == "" {
		s.logger.Warn("group has no email contact, skipping resource request notice", logFields...)
		return
	}

	if err := s.emailClient.Send(ctx, to, subject, body); err != nil {
		s.logger.Warn("failed to send resource request email", append(logFields, zap.Error(err))...)
	}
}

func (s *service) GetGroupResourceRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupResourceRequest, int, entities.PageScope, error) {
	reqs, pendingCount, ps, err := s.repo.GetGroupResourceRequestsByFaculty(ctx, faculty, status, pageScope)
	if err != nil {
		s.logger.Error("failed to get group resource requests by faculty", zap.Error(err))
		return nil, 0, entities.PageScope{}, err
	}
	return reqs, pendingCount, ps, nil
}

func (s *service) GetGroupResourceRequestsByGroupID(ctx context.Context, groupID string, status string, pageScope entities.PageScope) ([]entities.GroupResourceRequest, entities.PageScope, error) {
	reqs, ps, err := s.repo.GetGroupResourceRequestsByGroupID(ctx, groupID, status, pageScope)
	if err != nil {
		s.logger.Error("failed to get group resource requests by group id", zap.Error(err))
		return nil, entities.PageScope{}, err
	}
	return reqs, ps, nil
}

func (s *service) GetPendingGroupResourceRequestsCountByFaculty(ctx context.Context, faculty entities.Faculty) ([]FacultyPendingCount, error) {
	counts, err := s.repo.GetPendingGroupResourceRequestsCountByFaculty(ctx, faculty)
	if err != nil {
		s.logger.Error("failed to get pending group resource requests count by faculty", zap.Error(err))
		return nil, err
	}
	return counts, nil
}

func (s *service) GetGroupResourceRequestByID(ctx context.Context, reqID string) (entities.GroupResourceRequest, error) {
	return s.repo.GetGroupResourceRequestByID(ctx, reqID)
}
