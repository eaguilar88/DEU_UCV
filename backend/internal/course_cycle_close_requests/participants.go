package course_cycle_close_requests

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/xuri/excelize/v2"
)

const (
	// maxParticipants caps the rows of a participants file, which bounds the certificates job.
	maxParticipants = 2000

	participantsSheet = "Participantes"
	instructionsSheet = "Instrucciones"

	ParticipantsTemplateFileName    = "participantes_aprobados.xlsx"
	ParticipantsTemplateContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

// participantsHeaders is the exact header row of the participants file, shared by the template
// generator and the parser so they can't drift.
var participantsHeaders = []string{"Nombres", "Apellidos", "Cédula", "Correo"}

var cedulaPattern = regexp.MustCompile(`^[VE]-[0-9]{4,10}$`)

var accentReplacer = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// ParticipantsFileError is a problem with the participants file that the submitter can fix.
// Row is the 1-based spreadsheet row, or 0 when the problem is not tied to a row.
type ParticipantsFileError struct {
	Row int
	Msg string
}

func (e *ParticipantsFileError) Error() string {
	if e.Row > 0 {
		return fmt.Sprintf("archivo de participantes, fila %d: %s", e.Row, e.Msg)
	}
	return "archivo de participantes: " + e.Msg
}

func fileError(msg string) error {
	return &ParticipantsFileError{Msg: msg}
}

func rowError(row int, msg string) error {
	return &ParticipantsFileError{Row: row, Msg: msg}
}

// parseParticipants reads the approved participants from a participants file built on the
// template: first sheet, header row 1, data from row 2. Fully empty rows are skipped.
func parseParticipants(fileName string, r io.Reader) ([]entities.Participant, error) {
	if strings.ToLower(filepath.Ext(fileName)) != ".xlsx" {
		return nil, fileError("debe ser un archivo .xlsx generado a partir de la plantilla")
	}

	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fileError("no se pudo leer el archivo .xlsx")
	}
	defer f.Close() //nolint:errcheck

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fileError("el archivo no tiene hojas")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fileError("no se pudo leer la hoja de participantes")
	}

	if len(rows) == 0 || !isHeaderRow(rows[0]) {
		return nil, rowError(1, fmt.Sprintf("los encabezados deben ser exactamente: %s", strings.Join(participantsHeaders, ", ")))
	}

	participants := make([]entities.Participant, 0, len(rows)-1)
	seen := make(map[string]int, len(rows)-1)
	for i, cells := range rows[1:] {
		rowNum := i + 2
		values := make([]string, len(participantsHeaders))
		empty := true
		for j := range values {
			if j < len(cells) {
				values[j] = strings.TrimSpace(cells[j])
			}
			if values[j] != "" {
				empty = false
			}
		}
		if empty {
			continue
		}

		if len(participants) == maxParticipants {
			return nil, fileError(fmt.Sprintf("no puede tener más de %d participantes", maxParticipants))
		}

		p, err := toParticipant(rowNum, values)
		if err != nil {
			return nil, err
		}
		if firstRow, ok := seen[p.Document]; ok {
			return nil, rowError(rowNum, fmt.Sprintf("la cédula %s ya aparece en la fila %d", p.Document, firstRow))
		}
		seen[p.Document] = rowNum
		participants = append(participants, p)
	}

	if len(participants) == 0 {
		return nil, fileError("no contiene participantes")
	}
	return participants, nil
}

func toParticipant(row int, values []string) (entities.Participant, error) {
	firstName, lastName, rawCedula, email := values[0], values[1], values[2], values[3]
	if firstName == "" {
		return entities.Participant{}, rowError(row, "nombres vacíos")
	}
	if lastName == "" {
		return entities.Participant{}, rowError(row, "apellidos vacíos")
	}
	if rawCedula == "" {
		return entities.Participant{}, rowError(row, "cédula vacía")
	}
	cedula, ok := normalizeCedula(rawCedula)
	if !ok {
		return entities.Participant{}, rowError(row, fmt.Sprintf("cédula inválida %q", rawCedula))
	}
	if email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil || addr.Address != email {
			return entities.Participant{}, rowError(row, fmt.Sprintf("correo inválido %q", email))
		}
	}
	return entities.Participant{
		FirstName: firstName,
		LastName:  lastName,
		Document:  cedula,
		Email:     email,
	}, nil
}

// normalizeCedula strips dots, dashes and spaces, uppercases the V/E prefix (V when missing) and
// reports whether the result is a well-formed cédula, e.g. "v 12.345.678" -> "V-12345678".
func normalizeCedula(raw string) (string, bool) {
	s := strings.NewReplacer(".", "", "-", "", " ", "").Replace(strings.ToUpper(raw))
	prefix := "V"
	if strings.HasPrefix(s, "V") || strings.HasPrefix(s, "E") {
		prefix, s = s[:1], s[1:]
	}
	cedula := prefix + "-" + s
	return cedula, cedulaPattern.MatchString(cedula)
}

func isHeaderRow(cells []string) bool {
	if len(cells) < len(participantsHeaders) {
		return false
	}
	for i, h := range participantsHeaders {
		if normalizeHeader(cells[i]) != normalizeHeader(h) {
			return false
		}
	}
	for _, extra := range cells[len(participantsHeaders):] {
		if strings.TrimSpace(extra) != "" {
			return false
		}
	}
	return true
}

func normalizeHeader(s string) string {
	return accentReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// buildParticipantsTemplate generates the participants file template: a header row on the first
// sheet, the cédula column formatted as text, and a second sheet with instructions.
func buildParticipantsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close() //nolint:errcheck

	if err := f.SetSheetName(f.GetSheetName(0), participantsSheet); err != nil {
		return nil, err
	}

	for i, h := range participantsHeaders {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue(participantsSheet, cell, h); err != nil {
			return nil, err
		}
	}

	// Number format 49 is "@" (text): keeps Excel from dropping the V- prefix or turning the
	// cédula into a number.
	textStyle, err := f.NewStyle(&excelize.Style{NumFmt: 49})
	if err != nil {
		return nil, err
	}
	if err := f.SetColStyle(participantsSheet, "C", textStyle); err != nil {
		return nil, err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}},
	})
	if err != nil {
		return nil, err
	}
	if err := f.SetCellStyle(participantsSheet, "A1", "D1", headerStyle); err != nil {
		return nil, err
	}

	widths := map[string]float64{"A": 28, "B": 28, "C": 18, "D": 34}
	for col, width := range widths {
		if err := f.SetColWidth(participantsSheet, col, col, width); err != nil {
			return nil, err
		}
	}
	if err := f.SetPanes(participantsSheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	}); err != nil {
		return nil, err
	}

	if _, err := f.NewSheet(instructionsSheet); err != nil {
		return nil, err
	}
	instructions := []string{
		"Incluya solo a los participantes que aprobaron el curso: cada fila recibirá un certificado.",
		"Llene la hoja \"Participantes\" a partir de la fila 2, sin modificar los encabezados.",
		"Nombres, Apellidos y Cédula son obligatorios. Correo es opcional.",
		"Cédula: con prefijo V- o E- (por ejemplo V-12345678). Si se omite, se asume V-.",
		fmt.Sprintf("Máximo %d participantes por archivo.", maxParticipants),
	}
	for i, line := range instructions {
		if err := f.SetCellValue(instructionsSheet, fmt.Sprintf("A%d", i+1), line); err != nil {
			return nil, err
		}
	}
	if err := f.SetColWidth(instructionsSheet, "A", "A", 100); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
