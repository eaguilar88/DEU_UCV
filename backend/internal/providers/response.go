package providers

import "github.com/eaguilar88/deu/internal/entities"

type GetProviderResponse struct {
	ID         string        `json:"proveedor_id"`
	UserID     string        `json:"usuario_id"`
	Name       string        `json:"nombre_proveedor"`
	Bio        string        `json:"biografia,omitempty"`
	Internal   bool          `json:"interno"`
	Code       string        `json:"codigo_proveedor"`
	Files      ProviderFiles `json:"archivos"`
	Logo       string        `json:"provider_avatar_url,omitempty"`
	Type       string        `json:"tipo,omitempty"`
	ProfitType string        `json:"tipo_lucro,omitempty"`
	Contact    []string      `json:"emails_contacto,omitempty"`
	Phones     []string      `json:"telefonos_contacto,omitempty"`
	Webpage    string        `json:"sitio_web,omitempty"`
	Active     bool          `json:"activo"`
}

type ProviderFiles struct {
	CI      string   `json:"ci,omitempty"`
	RIF     string   `json:"rif,omitempty"`
	ISLR    string   `json:"islr,omitempty"`
	Resumes []string `json:"resumenes,omitempty"`
	Others  []string `json:"otros,omitempty"`
}

// providerToResponse converts a Provider entity to GetProviderResponse.
func providerToResponse(entity entities.Provider) GetProviderResponse {
	response := GetProviderResponse{
		ID:         entity.ID,
		UserID:     entity.User.ID,
		Name:       entity.User.FirstName + " " + entity.User.LastName,
		Internal:   entity.IsInternal,
		Code:       entity.Code,
		Active:     entity.DeletedAt == "",
		Bio:        entity.Bio,
		Type:       string(entity.Type),
		ProfitType: string(entity.ProfitType),
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

type CreateProviderResponse struct {
	ID   string `json:"id"`
	Code string `json:"codigo_proveedor"`
}

type GetProvidersResponse struct {
	Providers []GetProviderResponse `json:"proveedores"`
	Pages     entities.PageScope    `json:"paginas"`
}

// providersToResponse converts a slice of Provider entities to GetProvidersResponse.
func providersToResponse(providers []entities.Provider, pageScope entities.PageScope) GetProvidersResponse {
	var responseProviders []GetProviderResponse
	for _, provider := range providers {
		responseProviders = append(responseProviders, providerToResponse(provider))
	}
	return GetProvidersResponse{
		Providers: responseProviders,
		Pages:     pageScope,
	}
}
