package application

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eaguilar88/deu/pkg/config"
	kitHTTP "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gorilla/mux"
	"github.com/oklog/oklog/pkg/group"
)

type HealthChecker interface {
	Ping() error
}

// application represents the structure that contains all the components of the service
// it's composed by tools, clients, repositories, services and servers
type application struct {
	logger log.Logger
	config config.Server
	router *mux.Router
	// menuSvc menu.Service
}

const (
	docsSource          = "./docs/openapi/service.yaml"
	noVersionDefinedYet = "Version to be defined"
)

// New instances a new application
// The application contains all the related components that allow the execution of the service
func New(logger log.Logger) (*application, error) {
	var app application
	var err error

	// ctx := context.Background()

	// Build application tools
	app.logger = logger
	app.config, err = app.buildConfig()
	if err != nil {
		return nil, err
	}

	// Build application repositories

	// Build application services
	// app.menuSvc = app.buildMenuSvc()

	// Build HTTP and GRPC servers
	if err := app.buildHTTPRouter(); err != nil {
		return nil, err
	}

	return &app, nil
}

// Run executes the application using the servers
// this function starts the both HTTP and gRPC server
func (app *application) Run() error {
	var group group.Group

	// Configure HTTP
	svr := http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.HTTPPort),
		Handler: app.router,
	}
	group.Add(func() error {
		app.logger.Log("msg", fmt.Sprintf("Listening HTTP server on :%d", app.config.HTTPPort))
		return svr.ListenAndServe()
	}, func(err error) {
		app.logger.Log("msg", fmt.Sprintf("Error in listen and serv http: %v", err))
		_ = svr.Close()
	})

	// Run application
	if err := group.Run(); err != nil {
		app.logger.Log("msg", fmt.Sprintf("Exiting on error: %v", err))
		return err
	}

	return nil
}

// buildConfig builds the application config
func (app *application) buildConfig() (config.Server, error) {
	return config.Read(app.logger)
}

// buildRedisCartRepository builds the Redis Cart repository
// func (app *application) buildCartRepository() *cart_repository.RedisCartRepository {
// 	return cart_repository.NewRedisCartRepository(
// 		app.sentinelClient,
// 		app.config.Redis.KeyPrefix,
// 		app.config.Redis.TTL,
// 		app.logger,
// 	)
// }

// buildMenuSvc builds the menu service
// func (app *application) buildMenuSvc() *menu.MenuService {
// 	return menu.NewService(
// 		// app.cfaProductClient,
// 		// app.cfaVenueClient,
// 		// app.cfaVenueClient,
// 		// app.productCatalogClient,
// 		app.logger,
// 	)
// }

// buildHTTPRouter build the HTTP server
func (app *application) buildHTTPRouter() error {
	// app.router = mux.NewRouter()
	// app.router.Use(otelmux.Middleware(opentelemetry.GetServiceName()))
	// metrics.BootstrapWithMux(app.router)
	// app.router.Use(appMetrics.UsageMiddleware)

	// var (
	// 	menuEndpoints = menu.MakeEndpoints(app.menuSvc, app.logger)
	// )

	// jwtDecodeFunc, err := jwt.MakeDecodeJWTToClaims(
	// 	app.config.JWTEncryptionKey,
	// 	otherJWT.SigningMethodHS256,
	// 	[]jwt.AppetizeClaimsVersion{jwt.VersionExternalV1},
	// 	app.logger,
	// )
	// if err != nil {
	// 	return err
	// }

	// options := []kitHTTP.ServerOption{
	// 	kitHTTP.ServerErrorEncoder(transport.MakeHTTPErrorEncoder(app.logger)),
	// 	kitHTTP.ServerBefore(kitJWT.HTTPToContext()),
	// 	kitHTTP.ServerBefore(jwtDecodeFunc),
	// 	kitHTTP.ServerBefore(transport.AddAPIKeyToContext),
	// 	kitHTTP.ServerBefore(request_id.AddRequestIDToContextHTTP),
	// 	kitHTTP.ServerAfter(request_id.AddRequestIDsToHTTP),
	// }
	// app.addMenuEndpoints(menuEndpoints, options)

	return nil
}

// addMenuEndpoints adds the menu endpoints to the service
func (app *application) addMenuEndpoints(options []kitHTTP.ServerOption) {
	// app.addMenuWithModifiersEndpoints(endpoints, options)
	// app.addMenuDefaultEndpoints(endpoints, options)
}

// getHealthStatus controls the health status of a health checker service
func (app *application) getHealthStatus(checker HealthChecker, name string) func() (status, version string, err error) {
	return func() (status, version string, err error) {
		e := checker.Ping()
		status, version, err = app.processHealthResult(e, name)
		return status, version, err
	}
}

// processHealthResult processes the health result of a health checker service
func (app *application) processHealthResult(err error, product string) (string, string, error) {
	if err != nil {
		level.Error(app.logger).Log("source", product, "error", err)
		return strconv.Itoa(http.StatusServiceUnavailable), noVersionDefinedYet, err
	}
	return strconv.Itoa(http.StatusOK), noVersionDefinedYet, err
}
