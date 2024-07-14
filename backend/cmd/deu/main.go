package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"database/sql"

	"github.com/eaguilar88/deu/docs"
	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/config"
	"github.com/eaguilar88/deu/pkg/repository"
	"github.com/eaguilar88/deu/pkg/transport"
	"github.com/eaguilar88/deu/pkg/users"
	kitJWT "github.com/go-kit/kit/auth/jwt"
	kitHTTP "github.com/go-kit/kit/transport/http"
	"github.com/oklog/oklog/pkg/group"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	docsSource          = "./docs/openapi/service.yaml"
	noVersionDefinedYet = "Version to be defined"
)

func main() {
	// ctx := context.Background()
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	var g group.Group

	config, err := config.Read(logger)
	if err != nil {
		level.Error(logger).Log("error parsing configuration.")
		os.Exit(1)
	}

	postgres, err := mustConnectToDB(config.Database)
	if err != nil {
		level.Error(logger).Log("message", "error connecting to the db", "error", err)
		os.Exit(1)
	}
	defer postgres.Close()

	r := mux.NewRouter()
	// authService := auth.NewAuthService()
	// addDocsRoute(r, docsSource, logger)
	// addAuthRoutes(ctx, authService, r)

	repository := repository.NewRepository(postgres, config.FilePath, logger)
	userSvc := users.NewUsersService(repository, logger)
	userEndpoints := users.MakeEndpoints(userSvc, logger, nil)

	commonHTTPOptions := []kitHTTP.ServerOption{
		kitHTTP.ServerBefore(kitJWT.HTTPToContext()),
	}
	addUserRoutes(r, userEndpoints, commonHTTPOptions)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HTTPPort),
		Handler: r,
	}
	{
		g.Add(func() error {
			fmt.Printf("Server listening in port: %d\n", config.HTTPPort)
			return srv.ListenAndServe()
		}, func(err error) {
			level.Error(logger).Log("error", err, "message", "error booting up the server. closing connection")
			srv.Close()
		})
	}
	if err = g.Run(); err != nil {
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

func addUserRoutes(r *mux.Router, endpoints users.Endpoints, options []kitHTTP.ServerOption) {
	//Get User Endpoint
	getUserHandler := users.GetUserHandleHTTP(endpoints.GetUser, options)
	path := fmt.Sprintf(transport.FormatUsers, transport.ParamUserID)
	r.Methods(http.MethodGet).Path(path).Handler(getUserHandler)

	//Get Users Endpoint
	getUsersHandler := users.GetUsersHandleHTTP(endpoints.GetUsers, options)
	r.Methods(http.MethodGet).Path(transport.PathUsers).Handler(getUsersHandler)

	//Create User Endpoint
	createUserHandler := users.CreateUserHandleHTTP(endpoints.GetUsers, options)
	r.Methods(http.MethodPost).Path(transport.PathUsers).Handler(createUserHandler)

	//Update User Endpoint
	updateUserHandler := users.UpdateUserHandleHTTP(endpoints.GetUsers, options)
	path = fmt.Sprintf(transport.FormatUsers, transport.ParamUserID)
	r.Methods(http.MethodGet).Path(path).Handler(updateUserHandler)

	//Delete User Endpoint
	deleteUserHandler := users.DeleteUserHandleHTTP(endpoints.GetUsers, options)
	path = fmt.Sprintf(transport.FormatUsers, transport.ParamUserID)
	r.Methods(http.MethodGet).Path(path).Handler(deleteUserHandler)
}
