package groups

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const maxEnrichmentConcurrency = 10

type Repository interface {
	GetGroupByID(ctx context.Context, groupID string) (entities.ExtensionGroup, error)
	GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope) ([]entities.ExtensionGroup, entities.PageScope, error)
	GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error)
	GetGroupsSimple(ctx context.Context) ([]entities.ExtensionGroup, error)

	// Unit of Work: Atomic operations
	CreateGroupWithRequests(ctx context.Context, group entities.ExtensionGroup, requests []entities.GroupRequest) (int64, []entities.GroupMember, error)

	// Yearly renewal: the proposal and its approval requests, created atomically
	CreateGroupRenewal(ctx context.Context, renewal entities.GroupRenewal, requests []entities.GroupRequest) (int64, error)
	CancelGroupRenewal(ctx context.Context, renewalID string) error

	// Individual operations (for flexibility)
	DeleteGroup(ctx context.Context, groupID string) error

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	GetContactsByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) ([]entities.Contact, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error

	// Approvers notified of a new group request
	GetFacultyCoordinatorEmails(ctx context.Context, faculty entities.Faculty) ([]string, error)
	GetDEUAdminEmails(ctx context.Context) ([]string, error)
}

// MailClient defines the email sending operations required by the groups service.
type MailClient interface {
	SendTemplate(ctx context.Context, to string, tmpl email.Template, data any) error
}

type StorageClient interface {
	UploadFile(ctx context.Context, files []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetPresignedFileURL(ctx context.Context, objectKey string) (string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

type service struct {
	repo        Repository
	storage     StorageClient
	emailClient MailClient
	// renewalWindow is how long before its renewal date a group may submit its renewal.
	renewalWindow time.Duration
	now           func() time.Time
	log           *zap.Logger
}

func NewService(repository Repository, storage StorageClient, emailClient MailClient, renewalWindow time.Duration, logger *zap.Logger) Service {
	return &service{
		repo:          repository,
		storage:       storage,
		emailClient:   emailClient,
		renewalWindow: renewalWindow,
		now:           time.Now,
		log:           logger,
	}
}

// GetGroup returns the full group to viewers allowed to see its private data (see
// entities.Viewer.CanManageGroup) and the public-facing view to everyone else.
func (s *service) GetGroup(ctx context.Context, groupID string, viewer entities.Viewer) (entities.ExtensionGroup, error) {
	group, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return entities.ExtensionGroup{}, err
	}

	private := viewer.CanManageGroup(group)
	if !private {
		redactPrivate(&group)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		s.enrichGroupFiles(gctx, &group, private)
		return nil
	})
	g.Go(func() error {
		s.enrichGroupContacts(gctx, &group)
		return nil
	})
	if private {
		g.Go(func() error {
			s.enrichMemberFiles(gctx, group.Members)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.log.Error("failed to enrich group", zap.Error(err), zap.String("group_id", groupID))
	}

	return group, nil
}

// GetGroups decides visibility per group, so a list can mix full views (e.g. the
// viewer's own group) with public ones.
func (s *service) GetGroups(ctx context.Context, filter entities.GroupFilter, pageScope entities.PageScope, viewer entities.Viewer) ([]entities.ExtensionGroup, entities.PageScope, error) {
	groups, page, err := s.repo.GetGroups(ctx, restrictFilter(filter, viewer), pageScope)
	if err != nil {
		return nil, entities.PageScope{}, err
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxEnrichmentConcurrency)
	for i := range groups {
		private := viewer.CanManageGroup(groups[i])
		if !private {
			redactPrivate(&groups[i])
		}
		g.Go(func() error {
			s.enrichGroupFiles(gctx, &groups[i], private)
			s.enrichGroupContacts(gctx, &groups[i])
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.log.Error("failed to enrich groups", zap.Error(err))
	}

	return groups, page, nil
}

func (s *service) GetRandomActiveGroups(ctx context.Context, limit int) ([]entities.ExtensionGroup, error) {
	groups, err := s.repo.GetRandomActiveGroups(ctx, limit)
	if err != nil {
		return nil, err
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxRandomGroupsLimit)
	for i := range groups {
		redactPrivate(&groups[i])
		g.Go(func() error {
			s.enrichGroupFiles(gctx, &groups[i], false)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.log.Error("failed to enrich random active groups", zap.Error(err))
	}

	return groups, nil
}

func (s *service) GetGroupsSimple(ctx context.Context) ([]entities.ExtensionGroup, error) {
	return s.repo.GetGroupsSimple(ctx)
}

// restrictFilter limits non-admin viewers (anonymous callers included) to active,
// non-deleted groups: unapproved and deleted groups are only listed for root, deu_admin
// and faculty_admin, whatever the query params ask for.
func restrictFilter(filter entities.GroupFilter, viewer entities.Viewer) entities.GroupFilter {
	if viewer.IsGlobalAdmin() || viewer.IsFacultyAdmin() {
		return filter
	}
	active := true
	filter.Active = &active
	filter.Deleted = false
	return filter
}

// redactPrivate strips everything but the group's public-facing data. The response
// fields are omitempty, so cleared fields drop out of the JSON.
func redactPrivate(group *entities.ExtensionGroup) {
	group.Members = nil
	group.Project = nil
	group.Owner = nil
	group.Active = false
	group.UpdatedAt = ""
}

// enrichGroupFiles sets the public logo URL and, when includePrivate is set, a pre-signed
// URL for the project document. Private URLs are never generated for public views.
func (s *service) enrichGroupFiles(ctx context.Context, group *entities.ExtensionGroup, includePrivate bool) {
	filesMap, err := s.repo.GetFilesByOwner(ctx, group.ID, entities.OwnerTypeExtensionGroup)
	if err != nil {
		s.log.Error("failed to get group files", zap.Error(err), zap.String("group_id", group.ID))
		return
	}

	if logo := filesMap.GetSingleFile(entities.GroupFileTypeLogo); logo != nil {
		if url, err := s.storage.GetFileURL(ctx, logo.Key); err == nil {
			logo.URL = url
		}
		group.Logo = logo
	}

	if !includePrivate {
		return
	}
	if project := filesMap.GetSingleFile(entities.GroupFileTypeProject); project != nil {
		if url, err := s.storage.GetPresignedFileURL(ctx, project.Key); err == nil {
			project.URL = url
		}
		group.Project = project
	}
}

func (s *service) enrichGroupContacts(ctx context.Context, group *entities.ExtensionGroup) {
	contacts, err := s.repo.GetContactsByOwner(ctx, group.ID, entities.OwnerTypeExtensionGroup)
	if err != nil {
		s.log.Error("failed to get group contacts", zap.Error(err), zap.String("group_id", group.ID))
		return
	}

	for _, c := range contacts {
		switch c.Type {
		case entities.ContactTypeEmail:
			group.Email = c.Value
		case entities.ContactTypePhone:
			group.Phone = c.Value
		}
	}
}

func (s *service) enrichMemberFiles(ctx context.Context, members []entities.GroupMember) {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxEnrichmentConcurrency)
	for i := range members {
		g.Go(func() error {
			filesMap, err := s.repo.GetFilesByOwner(gctx, members[i].ID, entities.OwnerTypeGroupMember)
			if err != nil {
				s.log.Error("failed to get member files", zap.Error(err), zap.String("member_id", members[i].ID))
				return nil
			}

			if doc := filesMap.GetSingleFile(entities.GroupMemberFileTypeDocument); doc != nil {
				if url, err := s.storage.GetPresignedFileURL(gctx, doc.Key); err == nil {
					doc.URL = url
				}
				members[i].Document = doc
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		s.log.Error("failed to enrich member files", zap.Error(err))
	}
}

func (s *service) CreateGroup(ctx context.Context, group entities.ExtensionGroup) (int64, string, error) {
	userID := group.Owner.ID
	providerCode, err := entities.GenerateProviderCode(entities.GroupProviderType)
	if err != nil {
		s.log.Error("failed to generate provider code",
			zap.Error(err),
			zap.String("action", "generate_code"),
		)
		return -1, "", fmt.Errorf("failed to generate provider code: %w", err)
	}

	requests := approvalRequests(group, "New group creation request")

	// Create group and requests in a single transaction
	groupID, insertedMembers, err := s.repo.CreateGroupWithRequests(ctx, group, requests)
	if err != nil {
		s.log.Error("failed to create group with requests",
			zap.Error(err),
			zap.String("action", "create_group_with_requests"),
			zap.String("group_name", group.Name),
		)
		return -1, "", fmt.Errorf("failed to create group with requests: %w", err)
	}

	// Prepare files metadata
	commonMetadata := map[string]string{
		"group_owner":   group.ID,
		"group_id":      fmt.Sprintf("%d", groupID),
		"group_name":    group.Name,
		"provider_code": providerCode,
	}

	files := []*entities.File{
		makeFileEntityFromFilePointer(group.Logo, groupID, userID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeLogo, commonMetadata),
		makeFileEntityFromFilePointer(group.Project, groupID, userID, entities.OwnerTypeExtensionGroup, entities.GroupFileTypeProject, commonMetadata),
	}

	for i, m := range insertedMembers {
		if i < len(group.Members) && group.Members[i].Document != nil {
			files = append(files, makeMemberFileEntity(group.Members[i].Document, groupID, m.ID, userID, commonMetadata))
		}
	}

	validFiles := make([]*entities.File, 0, len(files))
	for _, f := range files {
		if f != nil {
			validFiles = append(validFiles, f)
		}
	}

	if len(validFiles) > 0 {
		if err := s.storage.UploadFile(ctx, validFiles); err != nil {
			s.log.Error("failed to upload group files",
				zap.Error(err),
				zap.String("group_id", fmt.Sprintf("%d", groupID)),
				zap.String("action", "upload_files"),
			)
			return -1, "", fmt.Errorf("failed to upload files: %w", err)
		}

		if err := s.repo.SaveFilesToDB(ctx, validFiles); err != nil {
			s.log.Error("failed to save files",
				zap.Error(err),
				zap.String("group_id", fmt.Sprintf("%d", groupID)),
				zap.String("action", "save_files"),
			)
			return -1, "", fmt.Errorf("failed to save files: %w", err)
		}
	}

	s.log.Info("group, requests, and files created successfully",
		zap.Int64("group_id", groupID),
		zap.Int("file_count", len(validFiles)))

	s.notifyApprovers(ctx, group, requests, email.TemplateGroupRequestSubmittedFaculty, email.TemplateGroupRequestSubmittedDEU)

	return groupID, providerCode, nil
}

// approvalRequests builds the requests that must all be approved for the group to be registered
// or renewed: only the DEU's for a multidisciplinary group, its faculty's and the DEU's otherwise.
func approvalRequests(group entities.ExtensionGroup, comments string) []entities.GroupRequest {
	base := entities.GroupRequest{
		Status:   entities.RequestStatus_UNDER_REVIEW,
		Comments: comments,
	}
	deuReq := base
	deuReq.Faculty = entities.FacultyDEU
	if isMultidisciplinary(group) {
		return []entities.GroupRequest{deuReq}
	}
	facultyReq := base
	facultyReq.Faculty = group.Faculty[0]
	return []entities.GroupRequest{facultyReq, deuReq}
}

// notifyApprovers tells the approvers of each of the group's requests that the group submitted
// them: the faculty's coordinators for a faculty request (facultyTmpl) and the DEU admins for the
// DEU one (deuTmpl). It is best-effort: the requests are already created, so failures are only
// logged.
func (s *service) notifyApprovers(ctx context.Context, group entities.ExtensionGroup, requests []entities.GroupRequest, facultyTmpl, deuTmpl email.Template) {
	faculty := multidisciplinaryLabel
	if !isMultidisciplinary(group) && len(group.Faculty) > 0 {
		faculty = string(group.Faculty[0])
	}
	data := email.GroupRequestSubmittedData{
		GroupName: group.Name,
		Faculty:   faculty,
	}

	for _, req := range requests {
		if req.Faculty == entities.FacultyDEU {
			admins, err := s.repo.GetDEUAdminEmails(ctx)
			if err != nil {
				s.log.Warn("failed to get DEU admins", zap.Error(err))
			}
			s.sendToAll(ctx, admins, deuTmpl, data)
			continue
		}

		coordinators, err := s.repo.GetFacultyCoordinatorEmails(ctx, req.Faculty)
		if err != nil {
			s.log.Warn("failed to get faculty coordinators", zap.Error(err), zap.String("faculty", string(req.Faculty)))
		}
		s.sendToAll(ctx, coordinators, facultyTmpl, data)
	}
}

func (s *service) sendToAll(ctx context.Context, recipients []string, tmpl email.Template, data any) {
	if len(recipients) == 0 {
		s.log.Warn("no recipients for notification", zap.String("template", string(tmpl)))
		return
	}
	for _, to := range recipients {
		if err := s.emailClient.SendTemplate(ctx, to, tmpl, data); err != nil {
			s.log.Warn("failed to send group request email", zap.Error(err), zap.String("template", string(tmpl)))
		}
	}
}

// multidisciplinaryLabel stands in for the faculty of a multidisciplinary group in emails.
const multidisciplinaryLabel = "Multidisciplinario"

func isMultidisciplinary(group entities.ExtensionGroup) bool {
	return group.IsMultidisciplinary || slices.Contains(group.Type, entities.MultidisciplinaryGroupType)
}

// RenewGroup submits the group's yearly renewal: group holds the whole group data, as on
// creation, and replaces the current data only once every approval request is approved. Only the
// group's user may submit it, from RenewalWindow before the renewal date on (or once overdue),
// and only one renewal may be under review at a time. It returns the renewal's ID.
func (s *service) RenewGroup(ctx context.Context, groupID string, group entities.ExtensionGroup) (int64, error) {
	current, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return -1, err
	}
	userID := group.Owner.ID
	if current.Owner == nil || current.Owner.ID != userID {
		return -1, ErrNotGroupOwner
	}
	if current.RenewalDueAt.IsZero() || s.now().Before(current.RenewalDueAt.Add(-s.renewalWindow)) {
		return -1, ErrRenewalNotOpen
	}

	group.ID = groupID
	requests := approvalRequests(group, "Group renewal request")
	renewalID, err := s.repo.CreateGroupRenewal(ctx, entities.GroupRenewal{
		GroupID:     groupID,
		Group:       group,
		SubmittedBy: userID,
	}, requests)
	if err != nil {
		s.log.Error("failed to create group renewal", zap.Error(err), zap.String("group_id", groupID))
		return -1, err
	}

	if err := s.storeRenewalFiles(ctx, group, renewalID, userID); err != nil {
		// Without its files the renewal can't be applied: drop it so it can be submitted again.
		if cancelErr := s.repo.CancelGroupRenewal(ctx, strconv.FormatInt(renewalID, 10)); cancelErr != nil {
			s.log.Error("failed to cancel group renewal without files", zap.Error(cancelErr), zap.Int64("renewal_id", renewalID))
		}
		return -1, err
	}

	s.log.Info("group renewal submitted", zap.String("group_id", groupID), zap.Int64("renewal_id", renewalID))
	s.notifyApprovers(ctx, group, requests, email.TemplateGroupRenewalSubmittedFaculty, email.TemplateGroupRenewalSubmittedDEU)

	return renewalID, nil
}

// storeRenewalFiles uploads the renewal's logo, project and member documents and records them as
// owned by the renewal. Approving the renewal hands them over to the group and its members.
func (s *service) storeRenewalFiles(ctx context.Context, group entities.ExtensionGroup, renewalID int64, userID string) error {
	prefix := fmt.Sprintf("files/groups/%s/renewals/%d", group.ID, renewalID)
	renewalIDStr := strconv.FormatInt(renewalID, 10)
	metadata := map[string]string{
		"group_id":   group.ID,
		"group_name": group.Name,
		"renewal_id": renewalIDStr,
	}

	files := make([]*entities.File, 0, len(group.Members)+2)
	add := func(file *entities.File, purpose string, public bool, extra map[string]string) {
		if file == nil {
			return
		}
		fileMetadata := make(map[string]string, len(metadata)+len(extra)+1)
		for k, v := range metadata {
			fileMetadata[k] = v
		}
		for k, v := range extra {
			fileMetadata[k] = v
		}
		fileMetadata["file_type"] = purpose
		file.OwnerID = renewalIDStr
		file.OwnerType = entities.OwnerTypeGroupRenewal
		file.Purpose = purpose
		file.Public = public
		file.MetaData = fileMetadata
		file.UploadedBy = userID
		files = append(files, file)
	}

	if group.Logo != nil {
		group.Logo.Key = fmt.Sprintf("%s/%s_%s", prefix, entities.GroupFileTypeLogo, group.Logo.Name)
	}
	add(group.Logo, entities.GroupFileTypeLogo, true, nil)
	if group.Project != nil {
		group.Project.Key = fmt.Sprintf("%s/%s_%s", prefix, entities.GroupFileTypeProject, group.Project.Name)
	}
	add(group.Project, entities.GroupFileTypeProject, false, nil)
	for i, m := range group.Members {
		if m.Document == nil {
			continue
		}
		m.Document.Key = fmt.Sprintf("%s/members/%d/%s_%s", prefix, i, entities.GroupMemberFileTypeDocument, m.Document.Name)
		add(m.Document, entities.GroupMemberFileTypeDocument, false, map[string]string{
			entities.GroupRenewalMemberIndexKey: strconv.Itoa(i),
		})
	}

	if len(files) == 0 {
		return nil
	}
	if err := s.storage.UploadFile(ctx, files); err != nil {
		s.log.Error("failed to upload group renewal files", zap.Error(err), zap.String("group_id", group.ID))
		return fmt.Errorf("failed to upload files: %w", err)
	}
	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.log.Error("failed to save group renewal files", zap.Error(err), zap.String("group_id", group.ID))
		return fmt.Errorf("failed to save files: %w", err)
	}
	return nil
}

func (s *service) DeleteGroup(ctx context.Context, groupID, userID string) error {
	return s.repo.DeleteGroup(ctx, groupID)
}

func makeFileEntityFromFilePointer(file *entities.File, groupID int64, uploadedBy string, ownerType entities.OwnerType, fileType string, metadata map[string]string) *entities.File {
	if file == nil {
		return nil
	}
	file.OwnerID = fmt.Sprintf("%d", groupID)
	file.OwnerType = ownerType

	fileMetadata := make(map[string]string)
	for k, v := range metadata {
		fileMetadata[k] = v
	}
	fileMetadata["file_type"] = fileType
	file.Key = fmt.Sprintf("files/groups/%d/%s_%s", groupID, fileType, file.Name)
	file.Public = fileType == entities.GroupFileTypeLogo
	file.MetaData = fileMetadata
	file.UploadedBy = uploadedBy
	return file
}

func makeMemberFileEntity(file *entities.File, groupID int64, memberID, uploadedBy string, metadata map[string]string) *entities.File {
	if file == nil {
		return nil
	}
	file.OwnerID = memberID
	file.OwnerType = entities.OwnerTypeGroupMember
	fileMetadata := make(map[string]string)
	for k, v := range metadata {
		fileMetadata[k] = v
	}
	fileMetadata["file_type"] = entities.GroupMemberFileTypeDocument
	fileMetadata["member_id"] = memberID
	file.Key = fmt.Sprintf("files/groups/%d/members/%s/%s_%s", groupID, memberID, entities.GroupMemberFileTypeDocument, file.Name)
	file.Public = false
	file.MetaData = fileMetadata
	file.UploadedBy = uploadedBy
	file.Purpose = entities.GroupMemberFileTypeDocument
	return file
}
