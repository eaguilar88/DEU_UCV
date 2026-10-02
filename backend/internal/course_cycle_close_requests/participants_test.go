package course_cycle_close_requests

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// buildParticipantsXLSX builds an .xlsx whose first sheet holds the given rows, starting at A1.
// A nil row leaves that spreadsheet row empty.
func buildParticipantsXLSX(rows ...[]string) []byte {
	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		for j, value := range row {
			cell, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				panic(err)
			}
			if err := f.SetCellStr(sheet, cell, value); err != nil {
				panic(err)
			}
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestParseParticipants(t *testing.T) {
	headers := participantsHeaders

	tests := []struct {
		name     string
		fileName string
		data     []byte
		want     []entities.Participant
		wantErr  string
	}{
		{
			name:     "valid rows with normalized cédulas, blank email and skipped empty row",
			fileName: "archivo_participantes.xlsx",
			data: buildParticipantsXLSX(
				headers,
				[]string{" Ana María ", "Pérez Gómez", "v-12.345.678", "ana@example.com"},
				nil,
				[]string{"Luis", "Rodríguez", "E 81234567", ""},
				[]string{"Carla", "Díaz", "9876543"},
			),
			want: []entities.Participant{
				{FirstName: "Ana María", LastName: "Pérez Gómez", Document: "V-12345678", Email: "ana@example.com"},
				{FirstName: "Luis", LastName: "Rodríguez", Document: "E-81234567"},
				{FirstName: "Carla", LastName: "Díaz", Document: "V-9876543"},
			},
		},
		{
			name:     "headers match ignoring case, accents and spaces",
			fileName: "lista.XLSX",
			data: buildParticipantsXLSX(
				[]string{" NOMBRES", "apellidos ", "Cedula", "CORREO"},
				[]string{"Ana", "Pérez", "V-12345678"},
			),
			want: []entities.Participant{{FirstName: "Ana", LastName: "Pérez", Document: "V-12345678"}},
		},
		{
			name:     "non-xlsx extension",
			fileName: "archivo_participantes.csv",
			data:     []byte("Nombres,Apellidos,Cédula,Correo\n"),
			wantErr:  "archivo de participantes: debe ser un archivo .xlsx generado a partir de la plantilla",
		},
		{
			name:     "not a spreadsheet",
			fileName: "archivo_participantes.xlsx",
			data:     []byte("not a zip"),
			wantErr:  "archivo de participantes: no se pudo leer el archivo .xlsx",
		},
		{
			name:     "wrong headers",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX([]string{"Nombre", "Cédula"}, []string{"Ana", "V-12345678"}),
			wantErr:  "archivo de participantes, fila 1: los encabezados deben ser exactamente: Nombres, Apellidos, Cédula, Correo",
		},
		{
			name:     "extra header column",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX([]string{"Nombres", "Apellidos", "Cédula", "Correo", "Nota"}),
			wantErr:  "archivo de participantes, fila 1: los encabezados deben ser exactamente: Nombres, Apellidos, Cédula, Correo",
		},
		{
			name:     "empty sheet",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(),
			wantErr:  "archivo de participantes, fila 1: los encabezados deben ser exactamente: Nombres, Apellidos, Cédula, Correo",
		},
		{
			name:     "headers only",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers),
			wantErr:  "archivo de participantes: no contiene participantes",
		},
		{
			name:     "blank first name",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "Pérez", "V-1234567"}, []string{"", "Díaz", "V-7654321"}),
			wantErr:  "archivo de participantes, fila 3: nombres vacíos",
		},
		{
			name:     "blank last name",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "", "V-1234567"}),
			wantErr:  "archivo de participantes, fila 2: apellidos vacíos",
		},
		{
			name:     "blank cédula",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "Pérez", "", "ana@example.com"}),
			wantErr:  "archivo de participantes, fila 2: cédula vacía",
		},
		{
			name:     "invalid cédula",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "Pérez", "X-12AB"}),
			wantErr:  `archivo de participantes, fila 2: cédula inválida "X-12AB"`,
		},
		{
			name:     "invalid email",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "Pérez", "V-1234567", "ana@"}),
			wantErr:  `archivo de participantes, fila 2: correo inválido "ana@"`,
		},
		{
			name:     "email with display name",
			fileName: "archivo_participantes.xlsx",
			data:     buildParticipantsXLSX(headers, []string{"Ana", "Pérez", "V-1234567", "Ana <ana@example.com>"}),
			wantErr:  `archivo de participantes, fila 2: correo inválido "Ana <ana@example.com>"`,
		},
		{
			name:     "duplicate cédula after normalization",
			fileName: "archivo_participantes.xlsx",
			data: buildParticipantsXLSX(headers,
				[]string{"Ana", "Pérez", "V-12345678"},
				[]string{"Luis", "Rodríguez", "V-7654321"},
				[]string{"Ana", "Pérez", "12.345.678"},
			),
			wantErr: "archivo de participantes, fila 4: la cédula V-12345678 ya aparece en la fila 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParticipants(tt.fileName, bytes.NewReader(tt.data))
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				var fileErr *ParticipantsFileError
				assert.True(t, errors.As(err, &fileErr), "error must be a ParticipantsFileError")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseParticipants_TooManyRows(t *testing.T) {
	rows := [][]string{participantsHeaders}
	for i := 0; i <= maxParticipants; i++ {
		rows = append(rows, []string{"Nombre", "Apellido", fmt.Sprintf("V-%d", 1000000+i)})
	}

	_, err := parseParticipants("archivo_participantes.xlsx", bytes.NewReader(buildParticipantsXLSX(rows...)))

	assert.EqualError(t, err, fmt.Sprintf("archivo de participantes: no puede tener más de %d participantes", maxParticipants))
}

func TestNormalizeCedula(t *testing.T) {
	tests := []struct {
		raw   string
		want  string
		valid bool
	}{
		{raw: "V-12345678", want: "V-12345678", valid: true},
		{raw: "v12.345.678", want: "V-12345678", valid: true},
		{raw: " e - 8.123.456 ", want: "E-8123456", valid: true},
		{raw: "12345678", want: "V-12345678", valid: true},
		{raw: "V-123", want: "V-123", valid: false},
		{raw: "P-12345678", want: "V-P12345678", valid: false},
		{raw: "V-12345678901", want: "V-12345678901", valid: false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, ok := normalizeCedula(tt.raw)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.valid, ok)
		})
	}
}

// TestBuildParticipantsTemplate_RoundTrip checks that a template filled in on its first sheet is
// accepted by the parser, and that the cédula column is formatted as text.
func TestBuildParticipantsTemplate_RoundTrip(t *testing.T) {
	data, err := buildParticipantsTemplate()
	require.NoError(t, err)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck

	assert.Equal(t, []string{participantsSheet, instructionsSheet}, f.GetSheetList())

	styleID, err := f.GetCellStyle(participantsSheet, "C2")
	require.NoError(t, err)
	style, err := f.GetStyle(styleID)
	require.NoError(t, err)
	assert.Equal(t, 49, style.NumFmt, "cédula column must be formatted as text")

	require.NoError(t, f.SetSheetRow(participantsSheet, "A2", &[]any{"Ana", "Pérez", "V-12345678", "ana@example.com"}))
	var filled bytes.Buffer
	require.NoError(t, f.Write(&filled))

	got, err := parseParticipants(ParticipantsTemplateFileName, &filled)
	require.NoError(t, err)
	assert.Equal(t, []entities.Participant{
		{FirstName: "Ana", LastName: "Pérez", Document: "V-12345678", Email: "ana@example.com"},
	}, got)
}
