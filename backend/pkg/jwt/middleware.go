package jwt

import (
	"context"
	"fmt"

	kitJWT "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	// Updated import
)

// JWTMiddleware checks for the existence and validity of a JWT in the context.
func JWTMiddleware(signer Signer, log log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			tokenString, ok := ctx.Value(kitJWT.JWTContextKey).(string)

			if !ok || tokenString == "" {
				level.Error(log).Log("message", "authorization token is missing")
				return nil, fmt.Errorf("authorization token is missing")
			}

			// Use your Signer to validate the token
			err := signer.ValidateToken(tokenString) // Pass token string
			if err != nil {
				level.Error(log).Log("message", "invalid authorization token", "error", err)
				return nil, fmt.Errorf("invalid authorization token: %v", err)
			}

			return next(ctx, request)
		}
	}
}
