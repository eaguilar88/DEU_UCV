package entities

import (
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
)

type ProviderPrefix string

type ProviderType string

const (
	CourseProvider     ProviderPrefix = "ECP"
	GroupProvider      ProviderPrefix = "GEX"
	CourseProviderType ProviderType   = "courses"
	GroupProviderType  ProviderType   = "groups"
)

type Provider struct {
	ID         string
	User       User
	Type       ProviderType
	IsInternal bool
	Code       string
	CI         *multipart.FileHeader
	RIF        *multipart.FileHeader
	ISLR       *multipart.FileHeader
	Resumes    []*multipart.FileHeader
	Others     []*multipart.FileHeader
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
