package auth

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	ID string `json:"id"`
}
