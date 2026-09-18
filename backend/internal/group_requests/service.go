package group_requests

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const passwordCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
	GetGroupRequestsByGroupID(ctx context.Context, groupID string) ([]entities.GroupRequest, error)
	GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error)
	CreateGroupAdminAndActivate(ctx context.Context, groupID string, adminUser entities.User) (int64, error)
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetUser(ctx context.Context, userID string) (*entities.User, error)
}

// MailClient defines the email sending operations required by the group_requests service.
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

func (s *service) ApproveGroupRequest(ctx context.Context, reqID string) error {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	if err := s.repo.ApproveGroupRequest(ctx, reqID); err != nil {
		return err
	}

	allRequests, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to fetch group requests for activation check", zap.Error(err), zap.String("group_id", req.GroupID))
		return err
	}

	allApproved := true
	for _, r := range allRequests {
		if r.ID == reqID {
			continue
		}
		if string(r.Status) != "approved" {
			allApproved = false
			break
		}
	}

	if allApproved {
		group, err := s.repo.GetGroupByID(ctx, req.GroupID)
		if err != nil {
			s.logger.Error("failed to fetch group before activation", zap.Error(err), zap.String("group_id", req.GroupID))
			return err
		}

		// The original requester, captured before CreateGroupAdminAndActivate below
		// reassigns extension_groups.user_id to the new dedicated group-admin login.
		owner, err := s.repo.GetUser(ctx, group.Owner.ID)
		if err != nil {
			s.logger.Error("failed to fetch original group requester", zap.Error(err), zap.String("group_id", req.GroupID))
			return err
		}

		s.logger.Info("all requests approved for group, creating group admin user and activating group",
			zap.String("group_id", req.GroupID),
			zap.String("original_owner_id", owner.ID),
			zap.String("original_owner_email", owner.Email),
		)

		rawPassword, err := generateRandomPassword(12)
		if err != nil {
			s.logger.Error("failed to generate password for new group admin", zap.Error(err))
			return err
		}

		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password for new group admin", zap.Error(err))
			return err
		}
		hashedPassword := string(hashedPasswordBytes)

		groupIDInt, err := strconv.Atoi(req.GroupID)
		if err != nil {
			return err
		}
		cleanGroupName := strings.ToLower(strings.ReplaceAll(req.GroupName, " ", "_"))

		adminUser := entities.User{
			CI:             fmt.Sprintf("%d", 99000000+groupIDInt),
			Email:          fmt.Sprintf("%s@extension.ucv.ve", cleanGroupName),
			FirstName:      "Representante",
			LastName:       req.GroupName,
			Password:       hashedPassword,
			DateOfBirth:    "2000-01-01",
			EducationLevel: "bachiller",
			Roles: []string{
				entities.RoleNameFromID(entities.RoleGroupAdmin),
			},
		}

		if _, err := s.repo.CreateGroupAdminAndActivate(ctx, req.GroupID, adminUser); err != nil {
			s.logger.Error("failed to create group admin user and activate group", zap.Error(err), zap.String("group_id", req.GroupID))
			return err
		}

		subject := "Datos de acceso del grupo de extensión"
		body := fmt.Sprintf(
			"El grupo %s ha sido aprobado y activado.\n\nUsuario: %s\nContraseña: %s\n\nPor favor inicie sesión y cambie su contraseña.",
			req.GroupName, adminUser.Email, rawPassword,
		)
		if err := s.emailClient.Send(ctx, owner.Email, subject, body); err != nil {
			s.logger.Error("failed to send group admin credentials email", zap.Error(err), zap.String("group_id", req.GroupID))
			return fmt.Errorf("error sending credentials email: %w", err)
		}
	}

	return nil
}

// generateRandomPassword returns a cryptographically random alphanumeric
// password of the given length.
func generateRandomPassword(length int) (string, error) {
	password := make([]byte, length)
	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordCharset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		password[i] = passwordCharset[n.Int64()]
	}

	return string(password), nil
}

func (s *service) RejectGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.RejectGroupRequest(ctx, reqID)
}

func (s *service) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error) {
	requests, ps, pendingCount, err := s.repo.GetGroupRequestsByFaculty(ctx, faculty, status, pageScope)
	if err != nil {
		s.logger.Error("failed to get group requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, 0, err
	}

	for i, req := range requests {
		approvals, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
		if err != nil {
			s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return nil, entities.PageScope{}, 0, err
		}
		requests[i].Approvals = approvals
	}

	return requests, ps, pendingCount, nil
}

func (s *service) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error) {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return entities.GroupRequest{}, err
	}

	req.Approvals, err = s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return entities.GroupRequest{}, err
	}

	return req, nil
}

func (s *service) GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error) {
	return s.repo.GetPendingGroupRequestsCounts(ctx, faculty)
}
