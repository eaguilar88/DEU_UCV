package providers

import "github.com/eaguilar88/deu/internal/entities"

type GetProviderResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func ProviderEntityToGetProviderResponse(entity entities.Provider) GetProviderResponse {
	return GetProviderResponse{
		ID: entity.ID,
	}
}

type GenProvidersResponse struct {
	Providers []GetProviderResponse `json:"providers"`
	Pages     entities.PageScope    `json:"pages"`
}

func ProvidersEntityToGetProvidersResponse(providers []entities.Provider, pageScope entities.PageScope) GenProvidersResponse {
	var responseProviders []GetProviderResponse
	for _, provider := range providers {
		responseProviders = append(responseProviders, ProviderEntityToGetProviderResponse(provider))
	}
	return GenProvidersResponse{
		Providers: responseProviders,
		Pages:     pageScope,
	}
}
