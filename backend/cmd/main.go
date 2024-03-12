package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/eaguilar88/deu/docs"
	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/config"
	"github.com/eaguilar88/deu/pkg/transport"

	"database/sql"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const docsSource = "./docs/openapi/service.yaml"

func main() {
	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stdout)

	config, err := config.Read(logger)
	if err != nil {
		level.Error(logger).Log("error parsing configuration.")
		os.Exit(1)
	}

	posgres, err := mustConnectToDB(config.Database)
	if err != nil {
		level.Error(logger).Log("message", "error connecting to the db", "error", err)
		os.Exit(1)
	}
	defer posgres.Close()

	r := mux.NewRouter()
	authService := auth.NewAuthService()
	addDocsRoute(r, docsSource, logger)
	addAuthRoutes(ctx, authService, r)

	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
	}

	fmt.Printf("Server listening in port: %d\n", config.HTTPPort)
	if err := srv.ListenAndServe(); err != nil {
		os.Exit(1)
	}
}

func addDocsRoute(r *mux.Router, docsRoute string, log log.Logger) {
	r.HandleFunc("/docs", docs.DocsHandler(r, docsRoute, log)).Methods("GET")
}

func mustConnectToDB(conf config.DatabaseConfig) (*sql.DB, error) {
	connection := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", conf.User, conf.Password, conf.Hostname, conf.Port, conf.Name)
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func addAuthRoutes(ctx context.Context, service *auth.AuthService, r *mux.Router) {
	r.HandleFunc("/health", transport.HealthHandler).Methods("GET")
	r.HandleFunc("/login", transport.LoginHandler(ctx, service)).Methods("POST")
}
