package auth

type LoginRequest struct {
	Username string `json:"usuario"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"usuario"`
	Password string `json:"password"`
}
