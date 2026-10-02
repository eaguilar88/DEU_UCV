package certificates

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	// renderConcurrency bounds the certificates rendered and uploaded at once.
	renderConcurrency = 4

	pdfContentType  = "application/pdf"
	zipContentType  = "application/zip"
	xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	participantsAttachmentName = "participantes.xlsx"
)

type Repository interface {
	GetCourseCycleCloseRequestByID(ctx context.Context, id string) (entities.CourseCycleCloseRequest, error)
	GetCourseCycleCloseRequestByCertificatesToken(ctx context.Context, token string) (entities.CourseCycleCloseRequest, error)
	GetCertificatesByCloseRequestID(ctx context.Context, closeRequestID int64) ([]entities.Certificate, error)
	GetCertificateByVerificationCode(ctx context.Context, code string) (entities.Certificate, error)
	MarkCertificateIssued(ctx context.Context, id int64, storageKey string) error
	ResetFailedCertificatesJob(ctx context.Context, closeRequestID int64) error

	GetUser(ctx context.Context, userID string) (*entities.User, error)
	// The cycle is closed once its close request is approved, and an issued certificate must keep
	// verifying if its course is later deactivated, so both are looked up including inactive ones.
	GetCoursePeriodByIDIncludingInactive(ctx context.Context, periodID string) (entities.CoursePeriod, error)
	GetCourseIncludingInactive(ctx context.Context, courseID string) (entities.Course, error)
}

// StorageClient defines the file storage operations required by the certificates service.
type StorageClient interface {
	PutObject(ctx context.Context, objectKey, contentType string, body io.Reader) error
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error)
}

// MailClient defines the email sending operations required by the certificates service.
type MailClient interface {
	SendTemplateWithAttachments(ctx context.Context, to string, tmpl email.Template, data any, attachments []email.Attachment) error
}

// Renderer converts an HTML document into a PDF.
type Renderer interface {
	Render(ctx context.Context, html []byte) ([]byte, error)
}

type service struct {
	repo          Repository
	storage       StorageClient
	emailClient   MailClient
	renderer      Renderer
	publicBaseURL string
	logger        *zap.Logger
}

// NewService builds the certificates service. publicBaseURL is where the verification pages are
// served (the host encoded in the certificates' QR codes).
func NewService(repo Repository, storage StorageClient, emailClient MailClient, renderer Renderer, publicBaseURL string, logger *zap.Logger) Service {
	return &service{
		repo:          repo,
		storage:       storage,
		emailClient:   emailClient,
		renderer:      renderer,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		logger:        logger,
	}
}

// GenerateForCloseRequest is the handler of entities.JobKindCourseCycleCertificates jobs. It
// renders and uploads every certificate not issued yet, builds the ZIP with all of them and emails
// the links to the user who submitted the close request. It is idempotent, so a retried job
// resumes where the previous attempt stopped.
func (s *service) GenerateForCloseRequest(ctx context.Context, payload []byte) error {
	var p entities.CourseCycleCertificatesPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("decoding certificates job payload: %w", err)
	}
	requestID := strconv.FormatInt(p.CloseRequestID, 10)
	logger := s.logger.With(zap.String("close_request_id", requestID))

	request, err := s.repo.GetCourseCycleCloseRequestByID(ctx, requestID)
	if err != nil {
		logger.Error("failed to get close request", zap.Error(err))
		return err
	}
	if request.Status != entities.RequestStatus_APPROVED || request.CertificatesToken == "" {
		logger.Error("close request is not approved", zap.String("status", string(request.Status)))
		return ErrInvalidCloseRequestStatus
	}

	certs, err := s.repo.GetCertificatesByCloseRequestID(ctx, request.ID)
	if err != nil {
		logger.Error("failed to get close request certificates", zap.Error(err))
		return err
	}
	if len(certs) == 0 {
		// Requests submitted before certificates existed have no participants on record.
		logger.Info("close request has no certificates to issue")
		return nil
	}

	period, course, err := s.courseOfCycle(ctx, request.CourseCycleID)
	if err != nil {
		logger.Error("failed to get close request course", zap.Error(err))
		return err
	}
	submitter, err := s.repo.GetUser(ctx, request.SubmittedByID)
	if err != nil {
		logger.Error("failed to get close request submitter", zap.Error(err))
		return err
	}

	if err := s.issuePending(ctx, certs, course, period); err != nil {
		logger.Error("failed to issue certificates", zap.Error(err))
		return err
	}

	zipKey := zipStorageKey(request.CourseCycleID, request.CertificatesToken)
	if err := s.uploadZip(ctx, zipKey, certs); err != nil {
		logger.Error("failed to build certificates zip", zap.Error(err))
		return err
	}

	if err := s.notifySubmitter(ctx, submitter.Email, course.Name, request.CertificatesToken, certs); err != nil {
		logger.Error("failed to send certificates email", zap.Error(err))
		return err
	}

	logger.Info("certificates delivered", zap.Int("count", len(certs)))
	return nil
}

// issuePending renders and uploads the certificates without a storage key, updating certs in place.
func (s *service) issuePending(ctx context.Context, certs []entities.Certificate, course entities.Course, period entities.CoursePeriod) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(renderConcurrency)
	for i := range certs {
		if certs[i].StorageKey != "" {
			continue
		}
		c := &certs[i]
		g.Go(func() error {
			return s.issue(gctx, c, course, period)
		})
	}
	return g.Wait()
}

func (s *service) issue(ctx context.Context, c *entities.Certificate, course entities.Course, period entities.CoursePeriod) error {
	issuedAt := time.Now()
	verifyURL := s.verificationURL(c.VerificationCode)
	qr, err := qrDataURI(verifyURL)
	if err != nil {
		return err
	}

	layout := layoutFor(course.Faculty)
	background, err := backgroundDataURI(layout.background)
	if err != nil {
		return err
	}
	var facultyName string
	if !layout.staticText {
		facultyName = course.Faculty.String()
	}
	endMonth, endYear := cycleMonthYear(period.EndDate)

	html, err := renderCertificate(certificateData{
		Background:       background,
		StaticText:       layout.staticText,
		FullName:         c.FullName(),
		Document:         c.Document,
		CourseName:       course.Name,
		Duration:         durationLabel(course.Duration),
		Modality:         modalityLabel(course.Location),
		EndMonth:         endMonth,
		EndYear:          endYear,
		Endorsement:      endorsement(course.Faculty),
		FacultyName:      facultyName,
		IssueDate:        spanishDate(issuedAt),
		VerificationCode: c.VerificationCode,
		VerifyHost:       hostOf(s.publicBaseURL),
		QRCode:           qr,
	})
	if err != nil {
		return err
	}

	pdf, err := s.renderer.Render(ctx, html)
	if err != nil {
		return fmt.Errorf("rendering certificate %d: %w", c.ID, err)
	}

	key := certificateStorageKey(c.CourseCycleID, c.VerificationCode)
	if err := s.storage.PutObject(ctx, key, pdfContentType, bytes.NewReader(pdf)); err != nil {
		return fmt.Errorf("uploading certificate %d: %w", c.ID, err)
	}
	if err := s.repo.MarkCertificateIssued(ctx, c.ID, key); err != nil {
		return fmt.Errorf("marking certificate %d as issued: %w", c.ID, err)
	}
	c.StorageKey = key
	c.IssuedAt = &issuedAt
	return nil
}

// uploadZip packs every certificate PDF into a ZIP and uploads it to key. The ZIP is written to a
// temporary file rather than memory, since a cohort has no upper bound on participants.
func (s *service) uploadZip(ctx context.Context, key string, certs []entities.Certificate) error {
	tmp, err := os.CreateTemp("", "certificados-*.zip")
	if err != nil {
		return err
	}
	defer func() {
		if err := tmp.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			s.logger.Warn("failed to close temporary zip", zap.Error(err))
		}
		if err := os.Remove(tmp.Name()); err != nil {
			s.logger.Warn("failed to remove temporary zip", zap.Error(err), zap.String("path", tmp.Name()))
		}
	}()

	zw := zip.NewWriter(tmp)
	for _, c := range certs {
		if err := s.addToZip(ctx, zw, c); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return s.storage.PutObject(ctx, key, zipContentType, tmp)
}

func (s *service) addToZip(ctx context.Context, zw *zip.Writer, c entities.Certificate) error {
	body, _, err := s.storage.GetObject(ctx, c.StorageKey)
	if err != nil {
		return fmt.Errorf("downloading certificate %d: %w", c.ID, err)
	}
	defer body.Close() //nolint:errcheck

	header := &zip.FileHeader{Name: zipEntryName(c), Method: zip.Store}
	if c.IssuedAt != nil {
		header.Modified = *c.IssuedAt
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, body)
	return err
}

func (s *service) notifySubmitter(ctx context.Context, to, courseName, token string, certs []entities.Certificate) error {
	links := make([]email.CertificateLink, 0, len(certs))
	for _, c := range certs {
		links = append(links, email.CertificateLink{Name: c.FullName(), URL: s.verificationURL(c.VerificationCode)})
	}

	sheet, err := s.participantsSheet(certs)
	if err != nil {
		return fmt.Errorf("building participants attachment: %w", err)
	}

	return s.emailClient.SendTemplateWithAttachments(ctx, to, email.TemplateCourseCycleCertificatesReady,
		email.CourseCycleCertificatesReadyData{
			CourseName:   courseName,
			ZipURL:       s.zipURL(token),
			Certificates: links,
		},
		[]email.Attachment{{Name: participantsAttachmentName, ContentType: xlsxContentType, Data: sheet}},
	)
}

// participantsSheet lists each participant with their certificate link, as an XLSX file.
func (s *service) participantsSheet(certs []entities.Certificate) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	sheet := f.GetSheetName(0)
	rows := [][]any{{"Nombres", "Apellidos", "Cédula", "Enlace"}}
	for _, c := range certs {
		rows = append(rows, []any{c.FirstName, c.LastName, c.Document, s.verificationURL(c.VerificationCode)})
	}
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		if err != nil {
			return nil, err
		}
		if err := f.SetSheetRow(sheet, cell, &row); err != nil {
			return nil, err
		}
	}
	bold, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(sheet, "A1", "D1", bold); err != nil {
		return nil, err
	}
	for col, width := range map[string]float64{"A": 28, "B": 28, "C": 16, "D": 70} {
		if err := f.SetColWidth(sheet, col, col, width); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// VerifyCertificate returns a valid certificate with its course. Unknown, unissued and revoked
// certificates all yield ErrCertificateNotFound, so the public page can't tell them apart.
func (s *service) VerifyCertificate(ctx context.Context, code string) (entities.CertificateDetails, error) {
	cert, err := s.validCertificate(ctx, code)
	if err != nil {
		return entities.CertificateDetails{}, err
	}
	period, course, err := s.courseOfCycle(ctx, cert.CourseCycleID)
	if err != nil {
		s.logger.Error("failed to get certificate course", zap.Error(err), zap.Int64("certificate_id", cert.ID))
		return entities.CertificateDetails{}, err
	}
	return entities.CertificateDetails{Certificate: cert, Course: course, Period: period}, nil
}

// GetCertificatePDF opens the PDF of a valid certificate.
func (s *service) GetCertificatePDF(ctx context.Context, code string) (io.ReadCloser, error) {
	cert, err := s.validCertificate(ctx, code)
	if err != nil {
		return nil, err
	}
	body, _, err := s.storage.GetObject(ctx, cert.StorageKey)
	if err != nil {
		s.logger.Error("failed to get certificate pdf", zap.Error(err), zap.Int64("certificate_id", cert.ID))
		return nil, err
	}
	return body, nil
}

// GetCertificatesZip opens the ZIP with every certificate of the close request owning token.
func (s *service) GetCertificatesZip(ctx context.Context, token string) (io.ReadCloser, error) {
	request, err := s.repo.GetCourseCycleCloseRequestByCertificatesToken(ctx, token)
	if err != nil {
		return nil, err
	}
	body, _, err := s.storage.GetObject(ctx, zipStorageKey(request.CourseCycleID, token))
	if err != nil {
		// The job has not uploaded it yet (or failed).
		s.logger.Warn("certificates zip not available", zap.Error(err), zap.Int64("close_request_id", request.ID))
		return nil, ErrCertificatesNotAvailable
	}
	return body, nil
}

// RetryCertificates requeues the failed certificates job of a close request.
func (s *service) RetryCertificates(ctx context.Context, closeRequestID string) error {
	id, err := strconv.ParseInt(closeRequestID, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid close request ID", ErrNoFailedCertificatesJob)
	}
	if err := s.repo.ResetFailedCertificatesJob(ctx, id); err != nil {
		if !errors.Is(err, ErrNoFailedCertificatesJob) {
			s.logger.Error("failed to reset certificates job", zap.Error(err), zap.String("close_request_id", closeRequestID))
		}
		return err
	}
	s.logger.Info("certificates job requeued", zap.String("close_request_id", closeRequestID))
	return nil
}

func (s *service) validCertificate(ctx context.Context, code string) (entities.Certificate, error) {
	cert, err := s.repo.GetCertificateByVerificationCode(ctx, code)
	if err != nil {
		if !errors.Is(err, ErrCertificateNotFound) {
			s.logger.Error("failed to get certificate", zap.Error(err))
		}
		return entities.Certificate{}, err
	}
	if !cert.IsValid() || cert.StorageKey == "" {
		return entities.Certificate{}, ErrCertificateNotFound
	}
	return cert, nil
}

func (s *service) courseOfCycle(ctx context.Context, cycleID int64) (entities.CoursePeriod, entities.Course, error) {
	period, err := s.repo.GetCoursePeriodByIDIncludingInactive(ctx, strconv.FormatInt(cycleID, 10))
	if err != nil {
		return entities.CoursePeriod{}, entities.Course{}, err
	}
	course, err := s.repo.GetCourseIncludingInactive(ctx, period.Course.ID)
	if err != nil {
		return entities.CoursePeriod{}, entities.Course{}, err
	}
	return period, course, nil
}

func (s *service) verificationURL(code string) string {
	return s.publicBaseURL + "/certificados/" + code
}

func (s *service) zipURL(token string) string {
	return s.publicBaseURL + "/certificados/lotes/" + token + "/zip"
}

func certificateStorageKey(cycleID int64, code string) string {
	return fmt.Sprintf("certificates/%d/%s.pdf", cycleID, code)
}

// zipStorageKey is the same on every attempt, so a retried job overwrites the ZIP.
func zipStorageKey(cycleID int64, token string) string {
	return fmt.Sprintf("certificates/%d/%s/certificados.zip", cycleID, token)
}

// zipEntryName names a certificate inside the ZIP, e.g. "Pérez Gómez, Ana María - V-12345678.pdf".
// The cédula keeps names unique within a close request.
func zipEntryName(c entities.Certificate) string {
	name := fmt.Sprintf("%s, %s - %s.pdf", c.LastName, c.FirstName, c.Document)
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r < ' ' {
			return '_'
		}
		return r
	}, name)
}
