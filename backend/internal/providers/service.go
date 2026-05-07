package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

// Custom domain errors
var (
	ErrProviderNotFound  = errors.New("provider not found")
	ErrInvalidProvider   = errors.New("invalid provider data")
	ErrProviderExists    = errors.New("provider already exists")
	ErrFileUploadFailed  = errors.New("failed to upload file")
	ErrFileNotFound      = errors.New("file not found")
	ErrNoIntentionLetter = errors.New("a new provider requires an intention letter")
)

// Repository defines the data access operations required by the providers service.
type Repository interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, code string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope, filters entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
	UpdateProviderStatus(ctx context.Context, providerID string) error
	CreateProviderRequest(ctx context.Context, providerID int64) error

	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetProviderContactInfo(ctx context.Context, providerID string) (entities.User, error)

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	GetFilesByOwnerAndPurpose(ctx context.Context, ownerID, ownerType, purpose string) ([]*entities.File, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error
	MarkCoursesWithDocumentation(ctx context.Context, providerID, from, to string) error
}

// MailClient defines the email sending operations required by the providers service.
type MailClient interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

// StorageClient defines the file storage operations required by the providers service.
type StorageClient interface {
	UploadFile(ctx context.Context, file []*entities.File) error
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
	GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

// service is the concrete implementation of the Service interface.
type service struct {
	repo        Repository
	storage     StorageClient
	emailClient MailClient
	logger      *zap.Logger
}

// NewService creates a new providers service with the given dependencies.
func NewService(repo Repository, storage StorageClient, emailClient MailClient, logger *zap.Logger) Service {
	return &service{
		repo:        repo,
		storage:     storage,
		emailClient: emailClient,
		logger:      logger,
	}
}

// GetProvider returns a provider by ID, including its associated files with pre-signed URLs.
func (s *service) GetProvider(ctx context.Context, providerID string) (entities.Provider, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.logger.Debug("getting provider",
		zap.String("provider_id", providerID),
		zap.String("action", "get_provider"),
	)

	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider by ID",
			zap.Error(err),
			zap.String("provider_id", providerID),
			zap.String("action", "get_provider"),
		)
		if errors.Is(err, ErrProviderNotFound) {
			return entities.Provider{}, ErrProviderNotFound
		}
		return entities.Provider{}, fmt.Errorf("failed to get provider by ID: %w", err)
	}

	files, err := s.getFilesForProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider files",
			zap.Error(err),
			zap.String("provider_id", providerID),
			zap.String("action", "get_provider_files"),
		)
		return entities.Provider{}, fmt.Errorf("failed to get provider files: %w", err)
	}

	provider.Files = files

	s.logger.Debug("provider retrieved successfully",
		zap.String("provider_id", providerID),
		zap.String("action", "get_provider"),
	)

	return provider, nil
}

// GetProviderByCode returns a provider by its assigned DEU code, including associated files.
func (s *service) GetProviderByCode(ctx context.Context, code string) (entities.Provider, error) {
	provider, err := s.repo.GetProviderByCode(ctx, code)
	if err != nil {
		s.logger.Error("failed to get provider by code", zap.Error(err))
		return entities.Provider{}, err
	}
	files, err := s.getFilesForProvider(ctx, provider.ID)
	if err != nil {
		s.logger.Error("failed to get files for provider", zap.Error(err), zap.String("provider_id", provider.ID))
		return entities.Provider{}, err
	}
	provider.Files = files
	return provider, nil
}

// GetProviders returns a paginated list of providers. Files are fetched concurrently for each provider.
func (s *service) GetProviders(ctx context.Context, pageScope entities.PageScope, filters entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error) {
	providers, pageScope, err := s.repo.GetProviders(ctx, pageScope, filters)
	if err != nil {
		s.logger.Error("failed to get providers", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for i := range providers {
		idx := i
		eg.Go(func() error {
			files, err := s.getFilesForProvider(egCtx, providers[idx].ID)
			if err != nil {
				return fmt.Errorf("failed to get files for provider %s: %w", providers[idx].ID, err)
			}
			providers[idx].Files = files
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		s.logger.Error("failed to get files for one or more providers", zap.Error(err))
		return nil, entities.PageScope{}, err
	}

	return providers, pageScope, nil
}

// CreateProvider registers a new provider, uploads its required files to storage,
// saves file metadata to the database, creates a provider request for admin review,
// and sends a confirmation email. Returns the new provider's database ID.
func (s *service) CreateProvider(ctx context.Context, provider *entities.Provider) (int64, error) {
	provider.IsActive = false
	createdProviderID, err := s.repo.CreateProvider(ctx, *provider)
	if err != nil {
		s.logger.Error("failed to create provider", zap.Error(err))
		return -1, err
	}

	provider.ID = fmt.Sprintf("%d", createdProviderID)
	commonMetadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}
	files, err := prepareFilesSlice(provider, createdProviderID, commonMetadata)
	if err != nil {
		s.logger.Error("failed to prepare files map", zap.Error(err))
		return -1, err
	}
	if err = s.storage.UploadFile(ctx, files); err != nil {
		s.logger.Error("failed to upload files file", zap.Error(err))
		return -1, err
	}

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.logger.Error("failed to save file metadata to database",
			zap.Error(err),
			zap.String("action", "save_metadata"),
		)
		return -1, fmt.Errorf("failed to save file metadata: %w", err)
	}

	s.logger.Debug("successfully uploaded and saved files",
		zap.Int("file_count", len(files)),
		zap.String("action", "upload_and_save"),
	)

	if err := s.repo.CreateProviderRequest(ctx, createdProviderID); err != nil {
		s.logger.Error("failed to create provider request",
			zap.Error(err),
			zap.String("action", "create_provider_request"),
		)
		return -1, fmt.Errorf("failed to create provider request: %w", err)
	}

	contact, err := s.repo.GetProviderContactInfo(ctx, provider.ID)
	if err != nil {
		s.logger.Error("failed to get provider contact info", zap.Error(err))
		return -1, err
	}

	if err := s.emailClient.Send(ctx, contact.Email, "Solicitud de registro recibida", "Tu solicitud de registro como proveedor ha sido recibida y está bajo revisión. Recibirás una notificación cuando sea procesada."); err != nil {
		s.logger.Warn("failed to send provider registration email",
			zap.Error(err),
			zap.String("action", "send_email"),
		)
	}

	return createdProviderID, nil
}

// UpdateProvider updates an existing provider's data and re-uploads its associated files.
func (s *service) UpdateProvider(ctx context.Context, providerID string, provider *entities.Provider) error {
	err := s.repo.UpdateProvider(ctx, providerID, *provider)
	if err != nil {
		s.logger.Error("failed to update provider", zap.Error(err))
		return err
	}

	commonMetadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_code":    provider.Code,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}
	providerIDInt, err := strconv.ParseInt(providerID, 10, 64)
	if err != nil {
		s.logger.Error("invalid providerID", zap.Error(err))
		return err
	}
	files, err := prepareFilesSlice(provider, providerIDInt, commonMetadata)
	if err != nil {
		s.logger.Error("failed to prepare files map", zap.Error(err))
		return err
	}
	if err = s.storage.UploadFile(ctx, files); err != nil {
		s.logger.Error("failed to upload files file", zap.Error(err))
		return err
	}
	// Save metadata to database
	s.logger.Debug("saving file metadata to database",
		zap.Int("file_count", len(files)),
		zap.String("action", "save_metadata"),
	)

	if err := s.repo.SaveFilesToDB(ctx, files); err != nil {
		s.logger.Error("failed to save file metadata to database",
			zap.Error(err),
			zap.String("action", "save_metadata"),
		)
		return fmt.Errorf("failed to save file metadata: %w", err)
	}

	s.logger.Debug("successfully uploaded and saved files",
		zap.Int("file_count", len(files)),
		zap.String("action", "upload_and_save"),
	)
	return nil
}

// UploadProviderDocuments uploads an intention letter (optional) and a commitment letter (required)
// for the provider associated with the given user. It gates on the intention letter existence,
// versions the commitment letter, and marks provider courses as documented.
func (s *service) UploadProviderDocuments(ctx context.Context, userID string, intentionLetter, commitmentLetter *entities.File) error {
	provider, err := s.repo.GetProviderByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get provider by user ID", zap.Error(err), zap.String("user_id", userID))
		if errors.Is(err, ErrProviderNotFound) {
			return ErrProviderNotFound
		}
		return fmt.Errorf("failed to get provider: %w", err)
	}

	existingIntention, err := s.repo.GetFilesByOwnerAndPurpose(ctx, provider.ID, string(entities.OwnerTypeProvider), entities.ProviderFileTypeIntentionLetter)
	if err != nil {
		s.logger.Error("failed to check intention letter", zap.Error(err))
		return fmt.Errorf("failed to check intention letter: %w", err)
	}
	if len(existingIntention) == 0 && intentionLetter == nil {
		return ErrNoIntentionLetter
	}

	existingCommitment, err := s.repo.GetFilesByOwnerAndPurpose(ctx, provider.ID, string(entities.OwnerTypeProvider), entities.ProviderFileTypeCommitmentLetter)
	if err != nil {
		s.logger.Error("failed to check commitment letters", zap.Error(err))
		return fmt.Errorf("failed to check commitment letters: %w", err)
	}
	nextVersion := len(existingCommitment) + 1
	prevDate := ""
	if len(existingCommitment) > 0 {
		prevDate = existingCommitment[len(existingCommitment)-1].CreatedAt
	}

	providerIDInt, err := strconv.ParseInt(provider.ID, 10, 64)
	if err != nil {
		s.logger.Error("invalid provider ID", zap.Error(err))
		return fmt.Errorf("invalid provider ID: %w", err)
	}
	metadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_type":    string(provider.Type),
		"provider_user_id": userID,
	}

	var filesToUpload []*entities.File
	if intentionLetter != nil {
		intentionLetter.Purpose = entities.ProviderFileTypeIntentionLetter
		intentionLetter.Version = 1
		filesToUpload = append(filesToUpload, makeFileEntityFromFilePointer(intentionLetter, providerIDInt, userID, entities.OwnerTypeProvider, metadata))
	}

	ext := filepath.Ext(commitmentLetter.Name)
	base := strings.TrimSuffix(commitmentLetter.Name, ext)
	commitmentLetter.Name = fmt.Sprintf("%s_v%d%s", base, nextVersion, ext)
	commitmentLetter.Purpose = entities.ProviderFileTypeCommitmentLetter
	commitmentLetter.Version = nextVersion
	filesToUpload = append(filesToUpload, makeFileEntityFromFilePointer(commitmentLetter, providerIDInt, userID, entities.OwnerTypeProvider, metadata))

	if err := s.storage.UploadFile(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to upload provider documents", zap.Error(err))
		return fmt.Errorf("failed to upload documents: %w", err)
	}
	if err := s.repo.SaveFilesToDB(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to save provider documents to DB", zap.Error(err))
		return fmt.Errorf("failed to save documents: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	if err := s.repo.MarkCoursesWithDocumentation(ctx, provider.ID, prevDate, now); err != nil {
		s.logger.Error("failed to mark courses with documentation", zap.Error(err))
		return fmt.Errorf("failed to mark courses: %w", err)
	}

	s.logger.Info("provider documents uploaded successfully",
		zap.String("provider_id", provider.ID),
		zap.Int("commitment_version", nextVersion),
	)
	return nil
}

// DeleteProvider soft-deletes a provider by ID.
func (s *service) DeleteProvider(ctx context.Context, providerID string) error {
	err := s.repo.DeleteProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to delete provider", zap.Error(err))
		return err
	}
	return nil
}

// ApproveProvider changes the status of a provider to active
func (s *service) ApproveProvider(ctx context.Context, providerID string) error {
	if err := s.repo.UpdateProviderStatus(ctx, providerID); err != nil {
		s.logger.Error("error enabling provider", zap.Error(err))
	}
	return nil
}

// prepareFilesSlice builds the full list of file entities from a provider's file pointers,
// setting ownership, storage key, and metadata on each file.
func prepareFilesSlice(provider *entities.Provider, providerID int64, metadata map[string]string) ([]*entities.File, error) {
	filesArr := []*entities.File{
		makeFileEntityFromFilePointer(provider.Files.CI, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.Files.RIF, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
		makeFileEntityFromFilePointer(provider.Files.ISLR, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata),
	}

	for _, resume := range provider.Files.Resumes {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(resume, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	for _, other := range provider.Files.Others {
		filesArr = append(filesArr, makeFileEntityFromFilePointer(other, providerID, provider.User.ID, entities.OwnerTypeProvider, metadata))
	}
	return filesArr, nil
}

// makeFileEntityFromFilePointer populates ownership and storage fields on a file pointer in-place and returns it.
func makeFileEntityFromFilePointer(file *entities.File, providerID int64, uploadedBy string, ownerType entities.OwnerType, metadata map[string]string) *entities.File {
	file.OwnerID = fmt.Sprintf("%d", providerID)
	file.OwnerType = ownerType
	file.Key = fmt.Sprintf("files/providers/%d/%s", providerID, file.Name)
	file.Public = false
	file.MetaData = metadata
	file.UploadedBy = uploadedBy
	file.CreatedAt = time.Now().Format(time.RFC3339)
	return file
}

// getFilesForProvider retrieves and populates file URLs for a provider.
func (s *service) getFilesForProvider(ctx context.Context, providerID string) (entities.ProviderFiles, error) {
	s.logger.Debug("getting files for provider",
		zap.String("provider_id", providerID),
		zap.String("action", "get_files"),
	)

	files, err := s.repo.GetFilesByOwner(ctx, providerID, entities.OwnerTypeProvider)
	if err != nil {
		s.logger.Error("failed to get files by owner",
			zap.Error(err),
			zap.String("provider_id", providerID),
			zap.String("action", "get_files"),
		)
		if errors.Is(err, ErrFileNotFound) {
			return entities.ProviderFiles{}, ErrFileNotFound
		}
		return entities.ProviderFiles{}, fmt.Errorf("failed to get files by owner: %w", err)
	}

	allFiles := files.GetAllFiles()
	errChan := make(chan error, len(allFiles))
	var wgFiles sync.WaitGroup

	for _, file := range allFiles {
		if file == nil {
			continue
		}
		wgFiles.Add(1)
		go func(f *entities.File) {
			defer wgFiles.Done()

			// Create timeout context for URL fetch
			urlCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			url, err := s.storage.GetFileURL(urlCtx, f.Key)
			if err != nil {
				s.logger.Error("failed to get file URL",
					zap.Error(err),
					zap.String("provider_id", providerID),
					zap.String("file_key", f.Key),
					zap.String("action", "get_file_url"),
				)
				errChan <- fmt.Errorf("failed to get URL for file %s: %w", f.Key, err)
				return
			}
			f.URL = url
		}(file)
	}

	wgFiles.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		// Collect all errors
		var errMsgs []string
		for err := range errChan {
			errMsgs = append(errMsgs, err.Error())
		}
		return entities.ProviderFiles{}, fmt.Errorf("failed to get URLs for some files: %v", errMsgs)
	}

	result := entities.ProviderFiles{
		CI:                files.GetSingleFile(entities.ProviderFileTypeCI),
		RIF:               files.GetSingleFile(entities.ProviderFileTypeRIF),
		ISLR:              files.GetSingleFile(entities.ProviderFileTypeISLR),
		Resumes:           files.GetMultipleFiles(entities.ProviderFileTypeResume),
		Others:            files.GetMultipleFiles(entities.ProviderFileTypeOther),
		IntentionLetter:   files.GetSingleFile(entities.ProviderFileTypeIntentionLetter),
		CommitmentLetters: files.GetMultipleFiles(entities.ProviderFileTypeCommitmentLetter),
	}

	s.logger.Debug("files retrieved successfully",
		zap.String("provider_id", providerID),
		zap.String("action", "get_files"),
		zap.Int("total_files", len(allFiles)),
	)

	return result, nil
}
