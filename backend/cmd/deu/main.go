package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/config"
	"github.com/eaguilar88/deu/pkg/courses"
	"github.com/eaguilar88/deu/pkg/endorsements"
	"github.com/eaguilar88/deu/pkg/jwt"
	"github.com/eaguilar88/deu/pkg/repository"
	"github.com/eaguilar88/deu/pkg/transport"
	"github.com/eaguilar88/deu/pkg/users"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

// const (
// 	docsSource          = "./docs/openapi/service.yaml"
// 	noVersionDefinedYet = "Version to be defined"
// )

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

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

	signer := jwt.NewJWTSigner(config.JWTEncryptionKey, config.TTL, &logger)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	repository := repository.NewRepository(postgres, config.FilePath, logger)
	authService := auth.NewAuthService(repository, signer, logger)
	authEndpoints := auth.MakeAuthEndpointsHandler(authService, logger)

	addHealthRoute(e)
	userSvc := users.NewUsersService(repository, logger)
	userEndpoints := users.MakeUserEndpointsHandler(userSvc, logger)

	endorsementSvc := endorsements.NewEndorsementsService(repository, logger)
	endorsementEndpoints := endorsements.MakeEndorsementEndpointsHandler(endorsementSvc, logger)

	courseSvc := courses.NewCoursesService(repository, logger)
	courseEndpoints := courses.MakeCourseEndpointsHandler(courseSvc, logger)

	addAuthRoutes(e, authEndpoints)
	addUserRoutes(e, userEndpoints, signer, logger)
	addEndorsementRoutes(e, endorsementEndpoints)
	addCourseRoutes(e, courseEndpoints)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.HTTPPort)))
}

func addHealthRoute(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return transport.HealthHandler(c)
	})
}

func mustConnectToDB(conf config.DatabaseConfig) (*sql.DB, error) {
	connection := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", conf.User, conf.Password, conf.Hostname, conf.Port, conf.Name)
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func addAuthRoutes(e *echo.Echo, endpoints auth.AuthEndpointsHandler) {
	e.POST("/auth/login", endpoints.LoginHandleHTTP)
}

func addUserRoutes(e *echo.Echo, endpoints users.UserEndpointsHandler, signer jwt.Signer, logger log.Logger) {
	g := e.Group("/users")
	g.Use(jwt.JWTMiddleware(signer, logger))
	g.GET("/:id", endpoints.GetUser)
	g.GET("", endpoints.GetUsers)
	g.POST("", endpoints.CreateUser)
	g.PUT("/:id", endpoints.UpdateUser)
	g.DELETE("/:id", endpoints.DeleteUser)
}

func addEndorsementRoutes(e *echo.Echo, endpoints endorsements.EndorsementEndpointsHandler) {
	e.GET("/endorsements/:id", endpoints.GetEndorsement)
	e.GET("/endorsements", endpoints.GetEndorsements)
	e.POST("/endorsements", endpoints.CreateEndorsement)
	e.PUT("/endorsements/:id", endpoints.UpdateEndorsement)
	e.DELETE("/endorsements/:id", endpoints.DeleteEndorsement)
}

func addCourseRoutes(e *echo.Echo, endpoints courses.CourseEndpointsHandler) {
	e.GET("/courses/:id", endpoints.GetCourse)
	e.GET("/courses", endpoints.GetCourses)
	e.POST("/courses", endpoints.CreateCourse)
	e.PUT("/courses/:id", endpoints.UpdateCourse)
	e.DELETE("/courses/:id", endpoints.DeleteCourse)
}
