package certificates

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"html"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/eaguilar88/deu/internal/certificates/mocks"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

const (
	testBaseURL = "https://extension.example.com"
	testToken   = "TOKENTOKENTOKENTOKENTOKEN2"
	codeAna     = "ANAANAANAANAANAANAANAANAA2"
	codeLuis    = "LUISLUISLUISLUISLUISLUIS22"
)

type testMocks struct {
	repo     *mocks.MockRepository
	storage  *mocks.MockStorageClient
	mail     *mocks.MockMailClient
	renderer *mocks.MockRenderer
}

func newTestService(t *testing.T) (Service, testMocks) {
	m := testMocks{
		repo:     mocks.NewMockRepository(t),
		storage:  mocks.NewMockStorageClient(t),
		mail:     mocks.NewMockMailClient(t),
		renderer: mocks.NewMockRenderer(t),
	}
	return NewService(m.repo, m.storage, m.mail, m.renderer, testBaseURL+"/", zap.NewNop()), m
}

func approvedRequest() entities.CourseCycleCloseRequest {
	return entities.CourseCycleCloseRequest{
		ID:                1,
		CourseCycleID:     2,
		SubmittedByID:     "5",
		Status:            entities.RequestStatus_APPROVED,
		CertificatesToken: testToken,
	}
}

// pendingAndIssuedCertificates returns Ana, not issued yet, and Luis, issued by a previous attempt.
func pendingAndIssuedCertificates() []entities.Certificate {
	issuedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	return []entities.Certificate{
		{ID: 10, CourseCycleID: 2, FirstName: "Ana María", LastName: "Pérez Gómez", Document: "V-12345678", VerificationCode: codeAna},
		{ID: 11, CourseCycleID: 2, FirstName: "Luis", LastName: "Rodríguez", Document: "E-81234567", VerificationCode: codeLuis,
			StorageKey: "certificates/2/" + codeLuis + ".pdf", IssuedAt: &issuedAt},
	}
}

func expectCourseLookups(repo *mocks.MockRepository) {
	repo.EXPECT().GetCoursePeriodByIDIncludingInactive(mock.Anything, "2").
		Return(entities.CoursePeriod{ID: "2", Course: entities.Course{ID: "3"}, StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-15T00:00:00Z"}, nil)
	repo.EXPECT().GetCourseIncludingInactive(mock.Anything, "3").
		Return(entities.Course{ID: "3", Name: "Fotografía Digital", Duration: "40", Location: "online", Faculty: entities.FacultyArquitecturaUrbanismo}, nil)
}

func TestService_GenerateForCloseRequest(t *testing.T) {
	svc, m := newTestService(t)
	anaKey := "certificates/2/" + codeAna + ".pdf"
	luisKey := "certificates/2/" + codeLuis + ".pdf"
	zipKey := "certificates/2/" + testToken + "/certificados.zip"

	m.repo.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(approvedRequest(), nil)
	m.repo.EXPECT().GetCertificatesByCloseRequestID(mock.Anything, int64(1)).Return(pendingAndIssuedCertificates(), nil)
	expectCourseLookups(m.repo)
	m.repo.EXPECT().GetUser(mock.Anything, "5").Return(&entities.User{ID: "5", Email: "owner@example.com"}, nil)

	// Only Ana is rendered: Luis was issued by a previous attempt.
	var renderedHTML string
	m.renderer.EXPECT().Render(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, html []byte) ([]byte, error) {
		renderedHTML = string(html)
		return []byte("%PDF-ana"), nil
	}).Once()
	m.storage.EXPECT().PutObject(mock.Anything, anaKey, pdfContentType, mock.Anything).Return(nil).Once()
	m.repo.EXPECT().MarkCertificateIssued(mock.Anything, int64(10), anaKey).Return(nil).Once()

	m.storage.EXPECT().GetObject(mock.Anything, anaKey).Return(io.NopCloser(strings.NewReader("%PDF-ana")), pdfContentType, nil)
	m.storage.EXPECT().GetObject(mock.Anything, luisKey).Return(io.NopCloser(strings.NewReader("%PDF-luis")), pdfContentType, nil)
	var zipData []byte
	m.storage.EXPECT().PutObject(mock.Anything, zipKey, zipContentType, mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, body io.Reader) error {
			var err error
			zipData, err = io.ReadAll(body)
			return err
		}).Once()

	var sentData email.CourseCycleCertificatesReadyData
	var sentAttachments []email.Attachment
	m.mail.EXPECT().SendTemplateWithAttachments(mock.Anything, "owner@example.com", email.TemplateCourseCycleCertificatesReady, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ string, _ email.Template, data any, attachments []email.Attachment) error {
			sentData = data.(email.CourseCycleCertificatesReadyData)
			sentAttachments = attachments
			return nil
		}).Once()

	err := svc.GenerateForCloseRequest(context.Background(), []byte(`{"close_request_id":1}`))
	require.NoError(t, err)

	// Certificate HTML carries the participant, course, faculty design and QR code.
	assert.Contains(t, renderedHTML, "Ana María Pérez Gómez")
	assert.Contains(t, renderedHTML, "V-12345678")
	assert.Contains(t, renderedHTML, "Fotografía Digital")
	assert.Contains(t, renderedHTML, "con una duración de 40 horas académicas")
	assert.Contains(t, renderedHTML, "finalizado en el mes de julio de 2026")
	assert.Contains(t, renderedHTML, "en modalidad en línea")
	assert.Contains(t, renderedHTML, "la Facultad de Arquitectura y Urbanismo y la Dirección de Extensión Universitaria")
	assert.Contains(t, renderedHTML, `src="data:image/jpeg;base64,`)
	assert.Contains(t, renderedHTML, codeAna)
	assert.Contains(t, renderedHTML, `src="data:image/png;base64,`)
	assert.Contains(t, renderedHTML, "extension.example.com")

	// ZIP holds every certificate, issued now or before.
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	assert.ElementsMatch(t, []string{
		"Pérez Gómez, Ana María - V-12345678.pdf",
		"Rodríguez, Luis - E-81234567.pdf",
	}, names)

	// Email links point at the public verification pages.
	assert.Equal(t, "Fotografía Digital", sentData.CourseName)
	assert.Equal(t, testBaseURL+"/certificados/lotes/"+testToken+"/zip", sentData.ZipURL)
	assert.Equal(t, []email.CertificateLink{
		{Name: "Ana María Pérez Gómez", URL: testBaseURL + "/certificados/" + codeAna},
		{Name: "Luis Rodríguez", URL: testBaseURL + "/certificados/" + codeLuis},
	}, sentData.Certificates)

	require.Len(t, sentAttachments, 1)
	assert.Equal(t, "participantes.xlsx", sentAttachments[0].Name)
	assert.Equal(t, xlsxContentType, sentAttachments[0].ContentType)
	f, err := excelize.OpenReader(bytes.NewReader(sentAttachments[0].Data))
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck
	rows, err := f.GetRows(f.GetSheetName(0))
	require.NoError(t, err)
	assert.Equal(t, [][]string{
		{"Nombres", "Apellidos", "Cédula", "Enlace"},
		{"Ana María", "Pérez Gómez", "V-12345678", testBaseURL + "/certificados/" + codeAna},
		{"Luis", "Rodríguez", "E-81234567", testBaseURL + "/certificados/" + codeLuis},
	}, rows)
}

// The certificate uses the design of the faculty that approved the course request (course.Faculty),
// even when the close request goes to the faculty that first received it (course.OriginFaculty).
func TestService_IssueUsesApprovingFacultyLayout(t *testing.T) {
	period := entities.CoursePeriod{ID: "2", EndDate: "2026-07-15"}
	tests := map[string]struct {
		course         entities.Course
		wantBackground string
		wantText       []string
		notWantText    []string
	}{
		"redirected from Medicina and approved by Farmacia": {
			course:         entities.Course{Name: "Farmacología", Faculty: entities.FacultyFarmacia, OriginFaculty: entities.FacultyMedicina},
			wantBackground: "diploma_farmacia.jpeg",
			wantText:       []string{"Otorgado a:", "Por tu participación en", "Mercy Ospina", "Facultad de Farmacia"},
			notWantText:    []string{"Medicina"},
		},
		"approved by the DEU": {
			course:         entities.Course{Name: "Oratoria", Faculty: entities.FacultyDEU, OriginFaculty: entities.FacultyDEU},
			wantBackground: "certificado_deu.svg",
			wantText:       []string{"la Dirección de Extensión Universitaria, como parte"},
			// The DEU design already prints these.
			notWantText: []string{"Otorgado a:", "Por tu participación en", "Mercy Ospina", "Facultad de"},
		},
		"unknown faculty falls back to the DEU": {
			course:         entities.Course{Name: "Oratoria"},
			wantBackground: "certificado_deu.svg",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc, m := newTestService(t)
			var renderedHTML string
			m.renderer.EXPECT().Render(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, html []byte) ([]byte, error) {
				renderedHTML = string(html)
				return []byte("%PDF"), nil
			}).Once()
			m.storage.EXPECT().PutObject(mock.Anything, mock.Anything, pdfContentType, mock.Anything).Return(nil).Once()
			m.repo.EXPECT().MarkCertificateIssued(mock.Anything, int64(10), mock.Anything).Return(nil).Once()

			cert := entities.Certificate{ID: 10, CourseCycleID: 2, FirstName: "Ana", LastName: "Pérez", Document: "V-1234567", VerificationCode: codeAna}
			require.NoError(t, svc.(*service).issue(context.Background(), &cert, tc.course, period))

			background, err := backgroundDataURI(tc.wantBackground)
			require.NoError(t, err)
			// html/template escapes "+" in attributes as "&#43;".
			assert.Contains(t, html.UnescapeString(renderedHTML), `src="`+string(background)+`"`)
			for _, text := range tc.wantText {
				assert.Contains(t, renderedHTML, text)
			}
			for _, text := range tc.notWantText {
				assert.NotContains(t, renderedHTML, text)
			}
		})
	}
}

func TestLayoutsCoverAllFaculties(t *testing.T) {
	for faculty := range entities.ValidFaculties {
		layout, ok := layouts[faculty]
		if assert.True(t, ok, "no certificate layout for %s", faculty) {
			_, err := backgroundDataURI(layout.background)
			assert.NoError(t, err, faculty)
		}
	}
	assert.Equal(t, deuLayout, layoutFor(""))
	assert.Equal(t, deuLayout, layoutFor("Desconocida"))
}

func TestService_GenerateForCloseRequest_Errors(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		prepare func(m testMocks)
		wantErr string
	}{
		{
			name:    "invalid payload",
			payload: `not json`,
			wantErr: "decoding certificates job payload: invalid character 'o' in literal null (expecting 'u')",
		},
		{
			name:    "close request not approved",
			payload: `{"close_request_id":1}`,
			prepare: func(m testMocks) {
				r := approvedRequest()
				r.Status = entities.RequestStatus_UNDER_REVIEW
				m.repo.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(r, nil)
			},
			wantErr: ErrInvalidCloseRequestStatus.Error(),
		},
		{
			name:    "render failure stops before the zip and email",
			payload: `{"close_request_id":1}`,
			prepare: func(m testMocks) {
				m.repo.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(approvedRequest(), nil)
				m.repo.EXPECT().GetCertificatesByCloseRequestID(mock.Anything, int64(1)).Return(pendingAndIssuedCertificates(), nil)
				expectCourseLookups(m.repo)
				m.repo.EXPECT().GetUser(mock.Anything, "5").Return(&entities.User{ID: "5", Email: "owner@example.com"}, nil)
				m.renderer.EXPECT().Render(mock.Anything, mock.Anything).Return(nil, errors.New("gotenberg down"))
			},
			wantErr: "rendering certificate 10: gotenberg down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, m := newTestService(t)
			if tt.prepare != nil {
				tt.prepare(m)
			}
			err := svc.GenerateForCloseRequest(context.Background(), []byte(tt.payload))
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestService_GenerateForCloseRequest_NoCertificates(t *testing.T) {
	svc, m := newTestService(t)
	m.repo.EXPECT().GetCourseCycleCloseRequestByID(mock.Anything, "1").Return(approvedRequest(), nil)
	m.repo.EXPECT().GetCertificatesByCloseRequestID(mock.Anything, int64(1)).Return(nil, nil)

	assert.NoError(t, svc.GenerateForCloseRequest(context.Background(), []byte(`{"close_request_id":1}`)))
}

func TestService_VerifyCertificate(t *testing.T) {
	issuedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := issuedAt.Add(time.Hour)
	valid := entities.Certificate{ID: 1, CourseCycleID: 2, VerificationCode: codeAna, StorageKey: "k", IssuedAt: &issuedAt}

	tests := []struct {
		name    string
		cert    entities.Certificate
		repoErr error
		wantErr error
	}{
		{name: "valid", cert: valid},
		{name: "unknown", repoErr: ErrCertificateNotFound, wantErr: ErrCertificateNotFound},
		{name: "not issued yet", cert: entities.Certificate{ID: 1, VerificationCode: codeAna}, wantErr: ErrCertificateNotFound},
		{
			name: "revoked",
			cert: func() entities.Certificate {
				c := valid
				c.RevokedAt = &revokedAt
				return c
			}(),
			wantErr: ErrCertificateNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, m := newTestService(t)
			m.repo.EXPECT().GetCertificateByVerificationCode(mock.Anything, codeAna).Return(tt.cert, tt.repoErr)
			if tt.wantErr == nil {
				expectCourseLookups(m.repo)
			}

			got, err := svc.VerifyCertificate(context.Background(), codeAna)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.cert, got.Certificate)
			assert.Equal(t, "Fotografía Digital", got.Course.Name)
		})
	}
}

func TestService_GetCertificatesZip_NotAvailable(t *testing.T) {
	svc, m := newTestService(t)
	m.repo.EXPECT().GetCourseCycleCloseRequestByCertificatesToken(mock.Anything, testToken).Return(approvedRequest(), nil)
	m.storage.EXPECT().GetObject(mock.Anything, "certificates/2/"+testToken+"/certificados.zip").Return(nil, "", errors.New("NoSuchKey"))

	_, err := svc.GetCertificatesZip(context.Background(), testToken)

	assert.ErrorIs(t, err, ErrCertificatesNotAvailable)
}

func TestService_RetryCertificates(t *testing.T) {
	t.Run("requeues the failed job", func(t *testing.T) {
		svc, m := newTestService(t)
		m.repo.EXPECT().ResetFailedCertificatesJob(mock.Anything, int64(1)).Return(nil)
		assert.NoError(t, svc.RetryCertificates(context.Background(), "1"))
	})
	t.Run("no failed job", func(t *testing.T) {
		svc, m := newTestService(t)
		m.repo.EXPECT().ResetFailedCertificatesJob(mock.Anything, int64(1)).Return(ErrNoFailedCertificatesJob)
		assert.ErrorIs(t, svc.RetryCertificates(context.Background(), "1"), ErrNoFailedCertificatesJob)
	})
	t.Run("invalid ID", func(t *testing.T) {
		svc, _ := newTestService(t)
		assert.ErrorIs(t, svc.RetryCertificates(context.Background(), "abc"), ErrNoFailedCertificatesJob)
	})
}

func TestMaskDocument(t *testing.T) {
	tests := map[string]string{
		"V-12345678": "V-12.3**.**8",
		"E-8123456":  "E-8.12*.**6",
		"V-123456":   "V-123.**6",
		"12345678":   "12.3**.**8",
	}
	for in, want := range tests {
		assert.Equal(t, want, maskDocument(in), in)
	}
}

func TestFormatting(t *testing.T) {
	assert.Equal(t, "15/01/2026", cycleDate("2026-01-15T00:00:00Z"))
	assert.Equal(t, "15/01/2026", cycleDate("2026-01-15"))
	assert.Equal(t, "", cycleDate(""))
	assert.Equal(t, "", cycleDate("not a date"))

	// 02:00 UTC is still the previous day in Caracas (UTC-4).
	assert.Equal(t, "28 de septiembre de 2026", spanishDate(time.Date(2026, 9, 29, 2, 0, 0, 0, time.UTC)))

	assert.Equal(t, "Dirección de Extensión Universitaria", facultyLabel(entities.FacultyDEU))
	assert.Equal(t, "Facultad de Ingeniería", facultyLabel(entities.FacultyIngenieria))
	assert.Equal(t, "", facultyLabel(""))

	month, year := cycleMonthYear("2026-07-15T00:00:00Z")
	assert.Equal(t, "julio", month)
	assert.Equal(t, "2026", year)
	month, year = cycleMonthYear("")
	assert.Empty(t, month)
	assert.Empty(t, year)

	assert.Equal(t, "la Dirección de Extensión Universitaria", endorsement(entities.FacultyDEU))
	assert.Equal(t, "la Facultad de Medicina y la Dirección de Extensión Universitaria", endorsement(entities.FacultyMedicina))

	assert.Equal(t, "en línea", modalityLabel("online"))
	assert.Equal(t, "presencial", modalityLabel("in_site"))
	assert.Equal(t, "mixta", modalityLabel("mixed"))
	assert.Equal(t, "", modalityLabel(""))

	assert.Equal(t, "40 horas académicas", durationLabel(" 40 "))
	assert.Equal(t, "3 semanas", durationLabel("3 semanas"))
	assert.Equal(t, "", durationLabel(""))

	assert.Equal(t, "Rojas_Díaz, Ana - V-1234567.pdf",
		zipEntryName(entities.Certificate{FirstName: "Ana", LastName: "Rojas/Díaz", Document: "V-1234567"}))
}
