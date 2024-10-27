package jwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	jwt "github.com/golang-jwt/jwt/v4"
)

// CustomClaims defines the structure of the JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// JWTMiddleware is a Go Kit middleware that validates the JWT token
func JWTMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// Extract the JWT token from the Authorization header
		req := request.(*http.Request)
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			return nil, fmt.Errorf("authorization header missing")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, fmt.Errorf("invalid authorization header format")
		}

		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the key for validation (replace with your actual secret key)
			return []byte("your_secret_key"), nil
		})

		if err != nil {
			return nil, fmt.Errorf("invalid token: %v", err)
		}

		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			// Store the claims in the context
			ctx = context.WithValue(ctx, "userID", claims.UserID)
			return next(ctx, request)
		}

		return nil, fmt.Errorf("invalid token claims")
	}
}
