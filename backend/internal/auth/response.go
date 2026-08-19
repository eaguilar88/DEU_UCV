package auth

type LoginResponse struct {
	Token              string            `json:"token"`
	User               LoginUserResponse `json:"usuario"`
	Faculty            string            `json:"facultad,omitempty"`
	ProviderCode       string            `json:"codigoProveedor,omitempty"`
	GroupID            string            `json:"grupoId,omitempty"`
	GroupName          string            `json:"nombreGrupo,omitempty"`
	CourseProviderID   string            `json:"proveedorId,omitempty"`
	CourseProviderName string            `json:"nombreProveedor,omitempty"`
}

type RegisterResponse struct {
	ID string `json:"id"`
}

type LoginUserResponse struct {
	ID    string   `json:"id"`
	Name  string   `json:"nombre"`
	Roles []string `json:"roles"`
}
