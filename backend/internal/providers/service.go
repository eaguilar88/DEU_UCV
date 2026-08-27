package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

// Custom domain errors
var (
	ErrProviderNotFound       = errors.New("provider not found")
	ErrInvalidProvider        = errors.New("invalid provider data")
	ErrProviderExists         = errors.New("provider already exists")
	ErrFileUploadFailed       = errors.New("failed to upload file")
	ErrFileNotFound           = errors.New("file not found")
	ErrMissingInitialContract = errors.New("carta_intencion y carta_compromiso son requeridas para el contrato inicial")
	ErrMissingAddendum        = errors.New("adenda es requerida")
	ErrNoCoursesToCoverage    = errors.New("no hay cursos aprobados sin cobertura legal para este proveedor")
)

// Repository defines the data access operations required by the providers service.
type Repository interface {
	GetProvider(ctx context.Context, providerID string) (entities.Provider, error)
	GetProviderByCode(ctx context.Context, code string) (entities.Provider, error)
	GetProviders(ctx context.Context, pageScope entities.PageScope, filters entities.ProviderFilters) ([]entities.Provider, entities.PageScope, error)
	CreateProvider(ctx context.Context, provider entities.Provider) (int64, error)
	UpdateProvider(ctx context.Context, providerID string, provider entities.Provider) error
	DeleteProvider(ctx context.Context, providerID string) error
	ApproveProvider(ctx context.Context, providerID string) error
	RejectProvider(ctx context.Context, providerID string) error
	CreateProviderRequest(ctx context.Context, providerID int64) error
	UpdateUserRole(ctx context.Context, userID, fromRole, toRole string) error

	GetProviderByUserID(ctx context.Context, userID string) (entities.Provider, error)
	GetProviderContactInfo(ctx context.Context, providerID string) (entities.User, error)

	// Files
	GetFilesByOwner(ctx context.Context, ownerID string, ownerType entities.OwnerType) (entities.GroupedFiles, error)
	GetFilesByOwnerAndPurpose(ctx context.Context, ownerID, ownerType, purpose string) ([]*entities.File, error)
	SaveFilesToDB(ctx context.Context, file []*entities.File) error

	// Provider contracts
	HasInitialContract(ctx context.Context, providerID string) (bool, error)
	GetUncoveredCourseIDs(ctx context.Context, providerID string) ([]string, error)
	CreateProviderContract(ctx context.Context, contract entities.ProviderContract, coveredCourseIDs []string) (entities.ProviderContract, error)
	GetProviderContracts(ctx context.Context, providerID string) ([]entities.ProviderContract, error)
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

	contracts, err := s.repo.GetProviderContracts(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider contracts",
			zap.Error(err),
			zap.String("provider_id", providerID),
			zap.String("action", "get_provider_contracts"),
		)
		return entities.Provider{}, fmt.Errorf("failed to get provider contracts: %w", err)
	}
	provider.Contracts = contracts

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
	code, err := entities.GenerateProviderCode(provider.Type)
	if err != nil {
		s.logger.Error("failed to generate provider code", zap.Error(err))
		return -1, err
	}
	provider.Status = entities.ProviderStatusUnderReview
	provider.Code = code
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

// SubmitProviderContract submits a provider's legal contract or addendum, admin-driven
// by provider ID. Whether this is the initial contract or an addendum is inferred
// server-side from whether the provider already has an initial contract on file:
// the initial contract requires intentionLetter+commitmentLetter, an addendum requires
// addendum. Every submission automatically covers all of the provider's currently
// approved-but-uncovered courses (never a client-supplied course list) — an addendum
// covering zero courses is rejected with ErrNoCoursesToCoverage, since it would be a
// no-op; the initial contract may legitimately cover zero courses for a brand-new provider.
func (s *service) SubmitProviderContract(ctx context.Context, providerID string, intentionLetter, commitmentLetter, addendum *entities.File) (entities.ProviderContract, error) {
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		s.logger.Error("failed to get provider", zap.Error(err), zap.String("provider_id", providerID))
		if errors.Is(err, ErrProviderNotFound) {
			return entities.ProviderContract{}, ErrProviderNotFound
		}
		return entities.ProviderContract{}, fmt.Errorf("failed to get provider: %w", err)
	}

	hasInitial, err := s.repo.HasInitialContract(ctx, provider.ID)
	if err != nil {
		s.logger.Error("failed to check for existing initial contract", zap.Error(err))
		return entities.ProviderContract{}, fmt.Errorf("failed to check for existing initial contract: %w", err)
	}

	contractType := entities.ContractTypeInitial
	if hasInitial {
		contractType = entities.ContractTypeAddendum
	}

	if contractType == entities.ContractTypeInitial {
		if intentionLetter == nil || commitmentLetter == nil {
			return entities.ProviderContract{}, ErrMissingInitialContract
		}
	} else if addendum == nil {
		return entities.ProviderContract{}, ErrMissingAddendum
	}

	uncoveredCourseIDs, err := s.repo.GetUncoveredCourseIDs(ctx, provider.ID)
	if err != nil {
		s.logger.Error("failed to get uncovered course IDs", zap.Error(err))
		return entities.ProviderContract{}, fmt.Errorf("failed to get uncovered course IDs: %w", err)
	}
	if contractType == entities.ContractTypeAddendum && len(uncoveredCourseIDs) == 0 {
		return entities.ProviderContract{}, ErrNoCoursesToCoverage
	}

	contract, err := s.repo.CreateProviderContract(ctx, entities.ProviderContract{
		ProviderID: provider.ID,
		Type:       contractType,
	}, uncoveredCourseIDs)
	if err != nil {
		s.logger.Error("failed to create provider contract", zap.Error(err))
		return entities.ProviderContract{}, fmt.Errorf("failed to create provider contract: %w", err)
	}

	metadata := map[string]string{
		"provider_id":      provider.ID,
		"provider_type":    string(provider.Type),
		"provider_user_id": provider.User.ID,
	}

	var filesToUpload []*entities.File
	if contractType == entities.ContractTypeInitial {
		intentionLetter.Purpose = entities.ProviderFileTypeIntentionLetter
		commitmentLetter.Purpose = entities.ProviderFileTypeCommitmentLetter
		filesToUpload = append(filesToUpload,
			makeContractFileEntity(intentionLetter, contract.ID, provider.User.ID, metadata),
			makeContractFileEntity(commitmentLetter, contract.ID, provider.User.ID, metadata),
		)
	} else {
		addendum.Purpose = entities.ProviderFileTypeAddendum
		filesToUpload = append(filesToUpload, makeContractFileEntity(addendum, contract.ID, provider.User.ID, metadata))
	}

	if err := s.storage.UploadFile(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to upload provider contract documents", zap.Error(err))
		return entities.ProviderContract{}, fmt.Errorf("failed to upload documents: %w", err)
	}
	if err := s.repo.SaveFilesToDB(ctx, filesToUpload); err != nil {
		s.logger.Error("failed to save provider contract documents to DB", zap.Error(err))
		return entities.ProviderContract{}, fmt.Errorf("failed to save documents: %w", err)
	}

	contract.Files = filesToUpload
	s.logger.Info("provider contract submitted successfully",
		zap.String("provider_id", provider.ID),
		zap.String("contract_id", contract.ID),
		zap.String("contract_type", string(contractType)),
		zap.Int("covered_courses", len(uncoveredCourseIDs)),
	)
	return contract, nil
}

// makeContractFileEntity populates ownership and storage fields for a file owned by a
// specific provider contract/addendum submission (not the provider directly), mirroring
// makeFileEntityFromFilePointer's role for provider-level files.
func makeContractFileEntity(file *entities.File, contractID, uploadedBy string, metadata map[string]string) *entities.File {
	file.OwnerID = contractID
	file.OwnerType = entities.OwnerTypeProviderContract
	file.Key = fmt.Sprintf("files/provider-contracts/%s/%s", contractID, file.Name)
	file.Public = false
	file.MetaData = metadata
	file.UploadedBy = uploadedBy
	file.CreatedAt = time.Now().Format(time.RFC3339)
	return file
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

// ApproveProvider sets the provider status to approved and promotes the user role from visitante to coordinador.
func (s *service) ApproveProvider(ctx context.Context, providerID, userID string) error {
	if err := s.repo.ApproveProvider(ctx, providerID); err != nil {
		s.logger.Error("error approving provider", zap.Error(err))
		return err
	}
	if err := s.repo.UpdateUserRole(ctx, userID, entities.RoleNameFromID(entities.RoleVisitante), entities.RoleNameFromID(entities.RoleCoordinador)); err != nil {
		s.logger.Error("error updating user role after provider approval", zap.Error(err))
		return err
	}
	return nil
}

// RejectProvider changes the status of a provider to active
func (s *service) RejectProvider(ctx context.Context, providerID string) error {
	if err := s.repo.RejectProvider(ctx, providerID); err != nil {
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
