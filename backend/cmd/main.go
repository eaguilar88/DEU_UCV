package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/eaguilar88/deu/pkg/transport"

	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	r := mux.NewRouter()
	authService := NewAuthService()
	addAuthRoutes(ctx, "", r)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", 8080),
	}

	fmt.Printf("Server listening in port: %d\n", 8080)
	if err := srv.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}

func addAuthRoutes(ctx context.Context, service string, r *mux.Router) {
	r.HandleFunc("/health", transport.HealthHandler).Methods("GET")
	r.HandleFunc("/login", transport.LoginHandler(ctx, authService)).Methods("POST")
}
