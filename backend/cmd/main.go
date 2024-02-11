package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/transport"

	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	r := mux.NewRouter()
	authService := auth.NewAuthService()
	addAuthRoutes(ctx, authService, r)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", 8080),
	}

	fmt.Printf("Server listening in port: %d\n", 8080)
	if err := srv.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}

func addAuthRoutes(ctx context.Context, service *auth.AuthService, r *mux.Router) {
	r.HandleFunc("/health", transport.HealthHandler).Methods("GET")
	r.HandleFunc("/login", transport.LoginHandler(ctx, service)).Methods("POST")
}
