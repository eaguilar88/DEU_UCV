package transport

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eaguilar88/deu/pkg/auth"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func LoginHandler(ctx context.Context, service *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Login Completed")
	}
}
