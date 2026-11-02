package providers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/providers/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newProviderForm builds a provider registration request with every file the diplomados UI sends
// for a juridical provider, minus omitFiles.
func newProviderForm(t *testing.T, partyType string, omitFiles ...string) echo.Context {
	t.Helper()

	omit := make(map[string]bool, len(omitFiles))
	for _, f := range omitFiles {
		omit[f] = true
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fields := map[string]string{
		"tipo_proveedor": "courses",
		"tipo_persona":   partyType,
		"tipo_lucro":     "no_lucrativo",
		"nombre":         "Academia Ejemplo",
		"bio":            "Formación continua",
		"es_interno":     "false",
		"facultad":       "Ciencias",
	}
	for key, value := range fields {
		require.NoError(t, writer.WriteField(key, value))
	}

	files := []string{
		entities.ProviderFileTypeCI,
		entities.ProviderFileTypeRIF,
		entities.ProviderFileTypeISLR,
		entities.ProviderFileTypeTitle,
		entities.ProviderFileTypeMercantileReg,
		entities.ProviderFileTypeLogo,
		"resumes",
	}
	for _, key := range files {
		if omit[key] {
			continue
		}
		part, err := writer.CreateFormFile(key, key+".pdf")
		require.NoError(t, err)
		_, err = part.Write([]byte("contenido de prueba"))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/providers", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	return echo.New().NewContext(req, httptest.NewRecorder())
}

func TestMakeProviderFromRequest(t *testing.T) {
	tests := []struct {
		name          string
		partyType     string
		omitFiles     []string
		wantErr       string
		wantParty     entities.ProviderPartyType
		wantISLR      bool
		wantMercantil bool
	}{
		{
			name:          "juridica with every file",
			partyType:     "juridica",
			wantParty:     entities.PartyTypeJuridical,
			wantISLR:      true,
			wantMercantil: true,
		},
		{
			name:      "natural without ISLR nor mercantile registry",
			partyType: "natural",
			omitFiles: []string{entities.ProviderFileTypeISLR, entities.ProviderFileTypeMercantileReg},
			wantParty: entities.PartyTypeNatural,
		},
		{
			name:      "invalid party type",
			partyType: "juridical",
			wantErr:   "tipo_persona must be 'natural' or 'juridica'",
		},
		{
			name:      "missing titulo",
			partyType: "natural",
			omitFiles: []string{entities.ProviderFileTypeTitle},
			wantErr:   "titulo is required",
		},
		{
			name:      "juridical without mercantile registry",
			partyType: "juridica",
			omitFiles: []string{entities.ProviderFileTypeMercantileReg},
			wantErr:   "registro_mercantil is required for juridical providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProviderForm(t, tt.partyType, tt.omitFiles...)

			got, err := makeProviderFromRequest(ctx, "user-1", zap.NewNop())
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantParty, got.PartyType)
			assert.NotNil(t, got.Files.Title)
			assert.Equal(t, tt.wantISLR, got.Files.ISLR != nil)
			assert.Equal(t, tt.wantMercantil, got.Files.MercantileReg != nil)
		})
	}
}

func TestProviderToResponse_PartyTypeAndFiles(t *testing.T) {
	got := providerToResponse(entities.Provider{
		PartyType: entities.PartyTypeJuridical,
		Files: entities.ProviderFiles{
			Title:         &entities.File{URL: "https://files.example/titulo.pdf"},
			MercantileReg: &entities.File{URL: "https://files.example/registro.pdf"},
		},
	})

	assert.Equal(t, "juridica", got.PartyType)
	assert.Equal(t, "https://files.example/titulo.pdf", got.Files.Title)
	assert.Equal(t, "https://files.example/registro.pdf", got.Files.MercantileReg)
	assert.Empty(t, got.Files.ISLR)
}

func TestHandler_CreateProvider_DuplicateRequest(t *testing.T) {
	for _, svcErr := range []error{ErrProviderRequestPending, ErrProviderAlreadyApproved} {
		t.Run(svcErr.Error(), func(t *testing.T) {
			ctx := newProviderForm(t, "natural")
			ctx.Set("userID", "user-1")
			svc := mocks.NewMockService(t)
			svc.EXPECT().CreateProvider(mock.Anything, mock.AnythingOfType("*entities.Provider")).Return(int64(-1), svcErr)

			err := NewHandler(svc, zap.NewNop()).CreateProvider(ctx)

			want := httperrors.NewConflict(svcErr.Error())
			require.Error(t, err)
			assert.Equal(t, want.Error(), err.Error())
			assert.Equal(t, want.StatusCode(), err.(httperrors.CustomError).StatusCode())
		})
	}
}
