package providers

import "github.com/eaguilar88/deu/internal/entities"

type GetProviderRequest struct {
	ID string
}

type GetProvidersRequest struct {
	PageScope entities.PageScope
}
