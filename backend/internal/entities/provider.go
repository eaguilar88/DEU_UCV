package entities

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ProviderPrefix string

type ProviderType string

type ProviderPartyType string

type ProviderProfitType string

const (
	// Provider code prefixes
	CourseProvider ProviderPrefix = "ECP"
	GroupProvider  ProviderPrefix = "GEX"

	// Provider types
	CourseProviderType ProviderType = "courses"
	GroupProviderType  ProviderType = "groups"

	// File types for providers
	ProviderFileTypeCI     = "ci"
	ProviderFileTypeRIF    = "rif"
	ProviderFileTypeISLR   = "islr"
	ProviderFileTypeResume = "resume"
	ProviderFileTypeOther  = "other"
	ProviderFileTypeLogo   = "logo"

	// Legal identity types
	PartyTypeNatural   ProviderPartyType = "natural"
	PartyTypeJuridical ProviderPartyType = "juridico"

	// Profit types
	ProfitTypeLucrativo   ProviderProfitType = "lucrativo"
	ProfitTypeNoLucrativo ProviderProfitType = "no_lucrativo"
)

type ProviderFiles struct {
	Logo    *File
	CI      *File
	RIF     *File
	ISLR    *File
	Resumes []*File
	Others  []*File
}

type Provider struct {
	ID         string
	User       User
	Name       string // Provider's display name (e.g., business name for juridical providers)
	Type       ProviderType
	PartyType  ProviderPartyType
	ProfitType ProviderProfitType
	IsInternal bool
	Bio        string
	Code       string
	IsActive   bool
	Faculty    Faculty
	Files      ProviderFiles
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}

type ProviderFilters struct {
	Type          ProviderType
	PartyType     ProviderPartyType
	ProfitType    ProviderProfitType
	IsInternal    *bool
	IsActive      *bool
	Code          string
	CreatedAtFrom string
	CreatedAtTo   string
}

func GenerateProviderCode(providerType ProviderType) (string, error) {
	short := strings.ReplaceAll(uuid.New().String(), "-", "")[:6]
	switch providerType {
	case CourseProviderType:
		return fmt.Sprintf("%s-%s", CourseProvider, short), nil
	case GroupProviderType:
		return fmt.Sprintf("%s-%s", GroupProvider, short), nil
	}
	return "", fmt.Errorf("invalid provider type: %s", providerType)
}
