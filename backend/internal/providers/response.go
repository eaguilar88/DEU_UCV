package providers

import "github.com/eaguilar88/deu/internal/entities"

type GetProviderResponse struct {
	ID       string        `json:"provider_id"`
	User     ProviderUser  `json:"user"`
	Internal bool          `json:"internal"`
	Code     string        `json:"provider_code"`
	Files    ProviderFiles `json:"files"`
	Active   bool          `json:"active"`
}

type ProviderUser struct {
	ID   string `json:"user_id"`
	Name string `json:"name"`
}

type ProviderFiles struct {
	CI      string   `json:"ci,omitempty"`
	RIF     string   `json:"rif,omitempty"`
	ISLR    string   `json:"islr,omitempty"`
	Resumes []string `json:"resumes,omitempty"`
	Others  []string `json:"others,omitempty"`
}

func ProviderEntityToGetProviderResponse(entity entities.Provider) GetProviderResponse {
	response := GetProviderResponse{
		ID: entity.ID,
		User: ProviderUser{
			ID:   entity.User.ID,
			Name: entity.User.FirstName + " " + entity.User.LastName,
		},
		Internal: entity.IsInternal,
		Code:     entity.Code,
		Active:   entity.DeletedAt == "",
	}

	files := ProviderFiles{}
	if entity.Files.CI != nil {
		files.CI = entity.Files.CI.URL
	}
	if entity.Files.RIF != nil {
		files.RIF = entity.Files.RIF.URL
	}
	if entity.Files.ISLR != nil {
		files.ISLR = entity.Files.ISLR.URL
	}
	if entity.Files.Resumes != nil {
		for _, file := range entity.Files.Resumes {
			if file != nil {
				files.Resumes = append(files.Resumes, file.URL)
			}
		}
	}
	if entity.Files.Others != nil {
		for _, file := range entity.Files.Others {
			if file != nil {
				files.Others = append(files.Others, file.URL)
			}
		}
	}
	response.Files = files
	return response
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
