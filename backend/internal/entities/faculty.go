package entities

import (
	"fmt"
	"strings"
)

// Faculty represents a faculty in the UCV university system.
// It maps to the PostgreSQL enum 'faculty_enum'.
type Faculty string

// Available faculties in the UCV university system
const (
	FacultyAgronomia                  Faculty = "Agronomía"
	FacultyArquitecturaUrbanismo      Faculty = "Arquitectura y Urbanismo"
	FacultyCiencias                   Faculty = "Ciencias"
	FacultyCienciasEconomicasSociales Faculty = "Ciencias Económicas y Sociales"
	FacultyCienciasJuridicasPoliticas Faculty = "Ciencias Jurídicas y Políticas"
	FacultyCienciasVeterinarias       Faculty = "Ciencias Veterinarias"
	FacultyFarmacia                   Faculty = "Farmacia"
	FacultyHumanidadesEducacion       Faculty = "Humanidades y Educación"
	FacultyIngenieria                 Faculty = "Ingeniería"
	FacultyMedicina                   Faculty = "Medicina"
	FacultyOdontologia                Faculty = "Odontología"
	FacultyDEU                        Faculty = "DEU" // DEU (Dirección de Extensión Universitaria
)

// ValidFaculties contains all valid faculty values
var ValidFaculties = map[Faculty]bool{
	FacultyAgronomia:                  true,
	FacultyArquitecturaUrbanismo:      true,
	FacultyCiencias:                   true,
	FacultyCienciasEconomicasSociales: true,
	FacultyCienciasJuridicasPoliticas: true,
	FacultyCienciasVeterinarias:       true,
	FacultyFarmacia:                   true,
	FacultyHumanidadesEducacion:       true,
	FacultyIngenieria:                 true,
	FacultyMedicina:                   true,
	FacultyOdontologia:                true,
	FacultyDEU:                        true,
}

// IsValid checks if the Faculty value is valid
func (f Faculty) IsValid() bool {
	return ValidFaculties[f]
}

func (f Faculty) String() string {
	return string(f)
}

func FromString(s string) (Faculty, error) {
	normalizedFaculties := make(map[string]Faculty, len(ValidFaculties))
	for faculty := range ValidFaculties {
		normalized := normalizeFacultyString(string(faculty))
		normalizedFaculties[normalized] = faculty
	}
	s = normalizeFacultyString(s)
	if faculty, ok := normalizedFaculties[s]; ok {
		return faculty, nil
	}
	return "", fmt.Errorf("invalid faculty: %s", s)
}

func normalizeFacultyString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")

	// Replace Spanish accents
	replacements := map[string]string{
		"á": "a",
		"é": "e",
		"í": "i",
		"ó": "o",
		"ú": "u",
		"ñ": "n",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	return s
}
