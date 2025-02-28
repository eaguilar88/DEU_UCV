package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/config"
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

	// endorsementSvc := endorsements.NewEndorsementsService(repository, logger)
	// endorsementEndpoints := endorsements.MakeEndpoints(endorsementSvc, logger, nil)

	// courseSvc := courses.NewCoursesService(repository, logger)
	// courseEndpoints := courses.MakeEndpoints(courseSvc, logger, nil)

	addAuthRoutes(e, authEndpoints)
	addUserRoutes(e, userEndpoints)
	// addEndorsementRoutes(e, endorsementEndpoints)
	// addCourseRoutes(e, courseEndpoints)

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

func addUserRoutes(e *echo.Echo, endpoints users.UserEndpointsHandler) {
	e.GET("/users/:id", endpoints.GetUser)
	e.GET("/users", endpoints.GetUsers)
	e.POST("/users", endpoints.CreateUser)
	e.PUT("/users/:id", endpoints.UpdateUser)
	e.DELETE("/users/:id", endpoints.DeleteUser)
}

// func addEndorsementRoutes(e *echo.Echo, endpoints endorsements.Endpoints) {
// 	e.GET(fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID), func(c echo.Context) error {
// 		return transport.GetEndorsementHandleHTTP(c.Response().Writer, c.Request(), endpoints.GetEndorsement)
// 	})
// 	e.GET(transport.PathEndorsements, func(c echo.Context) error {
// 		return transport.GetEndorsementsHandleHTTP(c.Response().Writer, c.Request(), endpoints.GetEndorsements)
// 	})
// 	e.POST(transport.PathEndorsements, func(c echo.Context) error {
// 		return transport.CreateEndorsementHandleHTTP(c.Response().Writer, c.Request(), endpoints.CreateEndorsement)
// 	})
// 	e.PUT(fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID), func(c echo.Context) error {
// 		return transport.UpdateEndorsementHandleHTTP(c.Response().Writer, c.Request(), endpoints.UpdateEndorsement)
// 	})
// 	e.DELETE(fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID), func(c echo.Context) error {
// 		return transport.DeleteEndorsementHandleHTTP(c.Response().Writer, c.Request(), endpoints.DeleteEndorsement)
// 	})
// }

// func addCourseRoutes(e *echo.Echo, endpoints courses.Endpoints) {
// 	e.GET(fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID), func(c echo.Context) error {
// 		return courses.GetCourseHandleHTTP(c.Response().Writer, c.Request(), endpoints.GetCourse)
// 	})
// 	e.GET(transport.PathCourses, func(c echo.Context) error {
// 		return courses.GetCoursesHandleHTTP(c.Response().Writer, c.Request(), endpoints.GetCourses)
// 	})
// 	e.POST(transport.PathCourses, func(c echo.Context) error {
// 		return courses.CreateCourseHandleHTTP(c.Response().Writer, c.Request(), endpoints.CreateCourse)
// 	})
// 	e.PUT(fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID), func(c echo.Context) error {
// 		return courses.UpdateCourseHandleHTTP(c.Response().Writer, c.Request(), endpoints.UpdateCourse)
// 	})
// 	e.DELETE(fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID), func(c echo.Context) error {
// 		return courses.DeleteCourseHandleHTTP(c.Response().Writer, c.Request(), endpoints.DeleteCourse)
// 	})
// }
