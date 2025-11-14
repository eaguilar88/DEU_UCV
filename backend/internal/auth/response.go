package auth

type LoginResponse struct {
	Token string            `json:"token"`
	User  LoginUserResponse `json:"usuario"`
}

type RegisterResponse struct {
	ID string `json:"id"`
}

type LoginUserResponse struct {
	ID    string   `json:"id"`
	Name  string   `json:"nombre"`
	Roles []string `json:"roles"`
}
