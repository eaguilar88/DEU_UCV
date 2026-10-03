package group_requests

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const passwordCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID, reason string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
	GetGroupRequestsByGroupID(ctx context.Context, groupID string) ([]entities.GroupRequest, error)
	GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error)
	ApproveGroupRequestAndActivate(ctx context.Context, reqID, groupID string, adminUser entities.User, notify func() error) (int64, error)
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error)

	// Yearly renewals
	GetGroupRenewal(ctx context.Context, renewalID string) (entities.GroupRenewal, error)
	ApplyGroupRenewal(ctx context.Context, reqID string, renewal entities.GroupRenewal) error
	RejectGroupRenewal(ctx context.Context, reqID, renewalID, reason string) error
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
}

// MailClient defines the email sending operations required by the group_requests service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

// StorageClient signs the links to a renewal's proposed files for its reviewers.
type StorageClient interface {
	GetPresignedFileURL(ctx context.Context, objectKey string) (string, error)
}

type service struct {
	repo        Repository
	emailClient MailClient
	storage     StorageClient
	logger      *zap.Logger
}

func NewService(repo Repository, emailClient MailClient, storage StorageClient, logger *zap.Logger) Service {
	return &service{
		repo:        repo,
		emailClient: emailClient,
		storage:     storage,
		logger:      logger,
	}
}

func (s *service) ApproveGroupRequest(ctx context.Context, reqID string) error {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	allRequests, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to fetch group requests for activation check", zap.Error(err), zap.String("group_id", req.GroupID))
		return err
	}

	// Only the requests of the same round count: the group's creation, or this same renewal.
	allApproved := true
	for _, r := range sameRound(allRequests, req.RenewalID) {
		if r.ID == reqID {
			continue
		}
		if string(r.Status) != "approved" {
			allApproved = false
			break
		}
	}

	if !allApproved {
		return s.repo.ApproveGroupRequest(ctx, reqID)
	}

	if req.RenewalID != "" {
		return s.applyRenewal(ctx, req)
	}

	group, err := s.repo.GetGroupByID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to fetch group before activation", zap.Error(err), zap.String("group_id", req.GroupID))
		return err
	}

	// The original requester, captured before ApproveGroupRequestAndActivate below
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

	credentials := email.GroupAdminCredentialsData{
		GroupName: req.GroupName,
		Username:  adminUser.Email,
		Password:  rawPassword,
	}
	sendCredentials := func() error {
		if err := s.emailClient.SendTemplate(ctx, owner.Email, email.TemplateGroupAdminCredentials, credentials); err != nil {
			s.logger.Error("failed to send group admin credentials email", zap.Error(err), zap.String("group_id", req.GroupID))
			return fmt.Errorf("error sending credentials email: %w", err)
		}
		return nil
	}

	if _, err := s.repo.ApproveGroupRequestAndActivate(ctx, reqID, req.GroupID, adminUser, sendCredentials); err != nil {
		s.logger.Error("failed to approve group request and activate group", zap.Error(err), zap.String("group_id", req.GroupID))
		return err
	}

	return nil
}

// applyRenewal approves the renewal's last pending request, replacing the group's data with the
// renewal's, and lets the group know.
func (s *service) applyRenewal(ctx context.Context, req entities.GroupRequest) error {
	logFields := []zap.Field{zap.String("group_id", req.GroupID), zap.String("renewal_id", req.RenewalID)}
	renewal, err := s.repo.GetGroupRenewal(ctx, req.RenewalID)
	if err != nil {
		s.logger.Error("failed to get group renewal", append(logFields, zap.Error(err))...)
		return err
	}

	if err := s.repo.ApplyGroupRenewal(ctx, req.ID, renewal); err != nil {
		s.logger.Error("failed to apply group renewal", append(logFields, zap.Error(err))...)
		return err
	}
	s.logger.Info("group renewal approved and applied", logFields...)

	// The renewal is already applied, so a notification failure is logged but not returned.
	to := renewal.Group.Email
	if to == "" {
		if to, err = s.groupContactEmail(ctx, req.GroupID); err != nil {
			s.logger.Warn("failed to get group contact email for renewal approval notice", append(logFields, zap.Error(err))...)
			return nil
		}
	}
	body := email.GroupRequestApprovedData{GroupName: renewal.Group.Name}
	if err := s.emailClient.SendTemplate(ctx, to, email.TemplateGroupRenewalApproved, body); err != nil {
		s.logger.Warn("failed to send group renewal approval email", append(logFields, zap.Error(err))...)
	}
	return nil
}

// sameRound keeps the requests of one approval round: the group's creation (renewalID empty) or
// one of its renewals.
func sameRound(requests []entities.GroupRequest, renewalID string) []entities.GroupRequest {
	round := make([]entities.GroupRequest, 0, len(requests))
	for _, r := range requests {
		if r.RenewalID == renewalID {
			round = append(round, r)
		}
	}
	return round
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

func (s *service) RejectGroupRequest(ctx context.Context, reqID, reason string) error {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	// Rejecting one request of a renewal rejects the whole renewal, so the group can submit a new one.
	if req.RenewalID != "" {
		if err := s.repo.RejectGroupRenewal(ctx, reqID, req.RenewalID, reason); err != nil {
			return err
		}
	} else if err := s.repo.RejectGroupRequest(ctx, reqID, reason); err != nil {
		return err
	}

	// The rejection is already saved, so a notification failure is logged but not returned.
	logFields := []zap.Field{zap.String("group_id", req.GroupID), zap.String("request_id", reqID)}
	to, err := s.groupContactEmail(ctx, req.GroupID)
	if err != nil {
		s.logger.Warn("failed to get group contact email for rejection notice", append(logFields, zap.Error(err))...)
		return nil
	}

	body := email.GroupRequestRejectedData{
		GroupName: req.GroupName,
		Reason:    reason,
	}
	if err := s.emailClient.SendTemplate(ctx, to, email.TemplateGroupRequestRejected, body); err != nil {
		s.logger.Warn("failed to send group rejection email", append(logFields, zap.Error(err))...)
	}
	return nil
}

// groupContactEmail returns the email saved in deu.contacts when the group was registered.
func (s *service) groupContactEmail(ctx context.Context, groupID string) (string, error) {
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
		requests[i].Approvals = sameRound(approvals, req.RenewalID)
	}

	return requests, ps, pendingCount, nil
}

func (s *service) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error) {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return entities.GroupRequest{}, err
	}

	approvals, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return entities.GroupRequest{}, err
	}
	req.Approvals = sameRound(approvals, req.RenewalID)

	if req.RenewalID != "" {
		renewal, err := s.repo.GetGroupRenewal(ctx, req.RenewalID)
		if err != nil {
			s.logger.Error("failed to get group renewal", zap.Error(err), zap.String("renewal_id", req.RenewalID))
			return entities.GroupRequest{}, err
		}
		s.attachRenewalFiles(ctx, &renewal)
		req.Renewal = &renewal
	}

	return req, nil
}

// attachRenewalFiles sets the renewal's proposed logo, project and member documents, with links
// for its reviewers.
func (s *service) attachRenewalFiles(ctx context.Context, renewal *entities.GroupRenewal) {
	files, err := s.repo.GetFilesByOwner(ctx, renewal.ID, entities.OwnerTypeGroupRenewal)
	if err != nil {
		s.logger.Error("failed to get group renewal files", zap.Error(err), zap.String("renewal_id", renewal.ID))
		return
	}

	sign := func(f *entities.File) *entities.File {
		if f == nil {
			return nil
		}
		url, err := s.storage.GetPresignedFileURL(ctx, f.Key)
		if err != nil {
			s.logger.Warn("failed to sign group renewal file", zap.Error(err), zap.String("key", f.Key))
			return f
		}
		f.URL = url
		return f
	}

	renewal.Group.Logo = sign(files.GetSingleFile(entities.GroupFileTypeLogo))
	renewal.Group.Project = sign(files.GetSingleFile(entities.GroupFileTypeProject))
	for _, doc := range files.GetMultipleFiles(entities.GroupMemberFileTypeDocument) {
		i, err := strconv.Atoi(doc.MetaData[entities.GroupRenewalMemberIndexKey])
		if err != nil || i < 0 || i >= len(renewal.Group.Members) {
			continue
		}
		renewal.Group.Members[i].Document = sign(doc)
	}
}

func (s *service) GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error) {
	return s.repo.GetPendingGroupRequestsCounts(ctx, faculty)
}
