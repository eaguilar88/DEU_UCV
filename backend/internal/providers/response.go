package providers

import "github.com/eaguilar88/deu/internal/entities"

type GetProviderResponse struct {
	ID       string        `json:"proveedor_id"`
	User     ProviderUser  `json:"usuario"`
	Internal bool          `json:"interno"`
	Code     string        `json:"codigo_proveedor"`
	Files    ProviderFiles `json:"archivos"`
	Active   bool          `json:"activo"`
}

type ProviderUser struct {
	ID   string `json:"usuario_id"`
	Name string `json:"nombre"`
}

type ProviderFiles struct {
	CI      string   `json:"ci,omitempty"`
	RIF     string   `json:"rif,omitempty"`
	ISLR    string   `json:"islr,omitempty"`
	Resumes []string `json:"resumenes,omitempty"`
	Others  []string `json:"otros,omitempty"`
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
	Providers []GetProviderResponse `json:"proveedores"`
	Pages     entities.PageScope    `json:"paginas"`
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
