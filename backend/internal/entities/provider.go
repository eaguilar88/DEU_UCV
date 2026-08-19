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

type ProviderStatus string

const (
	// Provider code prefixes
	CourseProvider ProviderPrefix = "ECP"
	GroupProvider  ProviderPrefix = "GEX"

	// Provider types
	CourseProviderType ProviderType = "courses"
	GroupProviderType  ProviderType = "groups"

	// File types for providers
	ProviderFileTypeCI               = "ci"
	ProviderFileTypeRIF              = "rif"
	ProviderFileTypeISLR             = "islr"
	ProviderFileTypeResume           = "curriculums"
	ProviderFileTypeOther            = "otros"
	ProviderFileTypeLogo             = "logo"
	ProviderFileTypeIntentionLetter  = "carta_intencion"
	ProviderFileTypeCommitmentLetter = "carta_compromiso"

	// Legal identity types
	PartyTypeNatural   ProviderPartyType = "natural"
	PartyTypeJuridical ProviderPartyType = "juridico"

	// Profit types
	ProfitTypeLucrativo   ProviderProfitType = "lucrativo"
	ProfitTypeNoLucrativo ProviderProfitType = "no_lucrativo"

	// Provider statuses
	ProviderStatusUnderReview ProviderStatus = "under_review"
	ProviderStatusRejected    ProviderStatus = "rejected"
	ProviderStatusActive      ProviderStatus = "active"
	ProviderStatusInactive    ProviderStatus = "inactive"
)

type ProviderFiles struct {
	Logo              *File
	CI                *File
	RIF               *File
	ISLR              *File
	Resumes           []*File
	Others            []*File
	IntentionLetter   *File
	CommitmentLetters []*File
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
	Status     ProviderStatus
	Faculty    Faculty
	Files      ProviderFiles
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
}

type ProviderFilters struct {
	Faculty Faculty
	Type    ProviderType
	ProviderAdminFilters
}

type ProviderAdminFilters struct {
	PartyType     ProviderPartyType
	ProfitType    ProviderProfitType
	IsInternal    *bool
	Status        ProviderStatus
	Code          string
	IsAdmin       bool
	CreatedAtFrom string
	CreatedAtTo   string
}

// IsCourseProvider reports whether the provider's code identifies it as a
// course provider (ECP-prefixed), as opposed to a group provider (GEX-prefixed).
func (p Provider) IsCourseProvider() bool {
	return strings.HasPrefix(p.Code, string(CourseProvider)+"-")
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
