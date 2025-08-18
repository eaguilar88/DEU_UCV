package entities

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type ProviderPrefix string

type ProviderType string

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
)

type ProviderFiles struct {
	CI      *File
	RIF     *File
	ISLR    *File
	Resumes []*File
	Others  []*File
}

type Provider struct {
	ID         string
	User       User
	Type       ProviderType
	IsInternal bool
	Code       string
	Files      ProviderFiles
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
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
