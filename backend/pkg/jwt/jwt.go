package jwt

import (
	"fmt"
	"time"

	"github.com/go-kit/log"
	jwt "github.com/golang-jwt/jwt/v4"
)

const tokenIssuer = "deu"

// CustomClaims defines the structure of the JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// Signer defines the interface for an authorization service.
type Signer interface {
	ValidateToken(tokenString string) error // Takes token string
	GenerateJWT(userID string, roles []string) (string, error)
}

type JWTSigner struct {
	SigningKey string
	TTL        uint32
	Logger     *log.Logger
}

func NewJWTSigner(signingKey string, ttl uint32, logger *log.Logger) Signer {
	return &JWTSigner{
		SigningKey: signingKey,
		TTL:        ttl,
		Logger:     logger,
	}
}

func (s *JWTSigner) ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.SigningKey), nil // Your signing key here
	})

	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Claims are valid
		return nil // Token is valid
	}

	return fmt.Errorf("invalid token")
}

func (s *JWTSigner) GenerateJWT(userID string, roles []string) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": tokenIssuer,
		"exp": time.Now().Add(time.Second * time.Duration(s.TTL)).Unix(), // Set expiration using s.TTL
		"iat": time.Now().Unix(),
		"v1": map[string]interface{}{
			"roles":  roles,
			"userID": userID,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(s.SigningKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}

// // JWTMiddleware is a Go Kit middleware that validates the JWT token
// func JWTMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
// 		// Extract the JWT token from the Authorization header
// 		req := request.(*http.Request)
// 		authHeader := req.Header.Get("Authorization")
// 		if authHeader == "" {
// 			return nil, fmt.Errorf("authorization header missing")
// 		}

// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			return nil, fmt.Errorf("invalid authorization header format")
// 		}

// 		tokenString := parts[1]

// 		// Parse and validate the token
// 		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
// 			// Ensure the signing method is HMAC
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			// Return the key for validation (replace with your actual secret key)
// 			return []byte("your_secret_key"), nil
// 		})

// 		if err != nil {
// 			return nil, fmt.Errorf("invalid token: %v", err)
// 		}

// 		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
// 			// Store the claims in the context
// 			ctx = context.WithValue(ctx, "userID", claims.UserID)
// 			return next(ctx, request)
// 		}

// 		return nil, fmt.Errorf("invalid token claims")
// 	}
// }
