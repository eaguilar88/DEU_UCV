package auth

import "database/sql"

func NewAuthService() *AuthService {
	return &AuthService{}
}

type AuthService struct {
	db sql.DB
}
