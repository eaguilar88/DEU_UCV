package auth

type LoginResponse struct {
	Token string            `json:"token"`
	User  LoginUserResponse `json:"user"`
}

type RegisterResponse struct {
	ID string `json:"id"`
}

type LoginUserResponse struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}
