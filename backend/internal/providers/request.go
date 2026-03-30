package providers

import (
	"mime/multipart"

	"github.com/eaguilar88/deu/internal/entities"
)

type GetProviderRequest struct {
	ID string
}

type GetProvidersRequest struct {
	PageScope entities.PageScope
	Filters   entities.ProviderFilters
}

type UpdateProviderRequest struct {
	ID         string
	UserID     string
	Type       string
	IsInternal bool
	CI         *multipart.FileHeader
	RIF        *multipart.FileHeader
	ISLR       *multipart.FileHeader
	Resumes    []*multipart.FileHeader
	Others     []*multipart.FileHeader
}

type DeleteProviderRequest struct {
	ID string
}
