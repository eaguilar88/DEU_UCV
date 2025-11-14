package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/eaguilar88/deu/internal/auth"
	"github.com/eaguilar88/deu/internal/config"
	"github.com/eaguilar88/deu/internal/course_periods"
	"github.com/eaguilar88/deu/internal/course_requests"
	"github.com/eaguilar88/deu/internal/courses"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/jwt"
	repository "github.com/eaguilar88/deu/internal/postgres_repository"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/eaguilar88/deu/internal/storage"
	"github.com/eaguilar88/deu/internal/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
// docsSource          = "./docs/openapi/service.yaml"
// noVersionDefinedYet = "Version to be defined"
)

type RegisterAdminEndpoints func(g *echo.Group)

func main() {
	logger, err := config.NewLogger()
	if err != nil {
		os.Exit(1)
	}

	defer logger.Sync() //nolint: errcheck
	config, err := config.Read(logger)
	if err != nil {
		logger.Error("error parsing configuration.")
		os.Exit(1)
	}

	postgres, err := mustConnectToDB(config.Database)
	if err != nil {
		logger.Error("error connecting to the db", zap.Error(err))
		os.Exit(1)
	}

	defer postgres.Close()
	logger.Info("connected to the db", zap.String("db", config.Database.String()))
	if err := postgres.Ping(); err != nil {
		logger.Error("error pinging the db", zap.Error(err))
		os.Exit(1)
	}

	signer := jwt.NewJWTSigner(config.JWTEncryptionKey, config.TTL, logger)
	bbClient, err := storage.NewB2Client(
		config.BlackBlazeB2.BucketName,
		config.BlackBlazeB2.KeyID,
		config.BlackBlazeB2.ApplicationKey,
		config.BlackBlazeB2.Endpoint,
		config.BlackBlazeB2.Region,
		logger,
	)
	if err != nil {
		logger.Error("error creating b2 client", zap.Error(err))
		os.Exit(1)
	}

	repository := repository.NewRepository(postgres, config.FilePath, logger)
	authService := auth.NewAuthService(repository, signer, logger)
	authEndpoints := auth.MakeAuthEndpointsHandler(authService, logger)

	userSvc := users.NewUsersService(repository, logger)
	userEndpoints := users.MakeUserEndpointsHandler(userSvc, logger)

	providerService := providers.NewProvidersService(repository, bbClient, logger)
	providerEndpoints := providers.MakeProviderEndpointsHandler(providerService, logger)

	courseSvc := courses.NewCoursesService(repository, logger)
	courseEndpoints := courses.MakeCourseEndpointsHandler(courseSvc, logger)

	cpService := course_periods.NewCoursePeriodsService(repository, logger)
	cpEndpoints := course_periods.MakeCoursePeriodEndpointsHandler(cpService, logger)

	groupService := groups.NewGroupsService(repository, logger)
	groupEndpoints := groups.MakeGroupEndpointsHandler(groupService, logger)

	groupRequestService := group_requests.NewGroupRequestService(repository, logger)
	groupRequestEndpoints := group_requests.MakeGroupRequestEndpointsHandler(groupRequestService, logger)

	courseRequestService := course_requests.NewCourseRequestService(repository, logger)
	courseRequestEndpoints := course_requests.MakeCourseRequestEndpointsHandler(courseRequestService, logger)

	e := echo.New()
	e.Validator = security.NewCustomValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	middlewares := []echo.MiddlewareFunc{
		jwt.JWTMiddleware(signer, logger),
	}

	addHealthRoute(e)
	addAuthRoutes(e, authEndpoints)
	addUserRoutes(e, userEndpoints, middlewares...)
	addProviderRoutes(e, providerEndpoints, middlewares...)
	addCourseRoutes(e, courseEndpoints, middlewares...)
	addCoursePeriodRoutes(e, cpEndpoints, middlewares...)
	addGroupsRoutes(e, groupEndpoints, middlewares...)

	addAdminRoutes(e, middlewares,
		groupRequestEndpoints.RegisterGroupRequestAdminEndpoints,
		courseRequestEndpoints.RegisterCourseRequestAdminEndpoints,
	)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.HTTPPort)))
}

func addHealthRoute(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok")
	})
}

func mustConnectToDB(conf config.DatabaseConfig) (*sql.DB, error) {
	connection := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Hostname,
		conf.Port,
		conf.Name,
	)
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func addAuthRoutes(e *echo.Echo, endpoints auth.AuthEndpointsHandler) {
	e.POST("/auth/login", endpoints.LoginHandleHTTP)
}

func addAdminRoutes(e *echo.Echo, middlewares []echo.MiddlewareFunc, handlers ...RegisterAdminEndpoints) {
	g := e.Group("/admin", middlewares...)
	for _, handler := range handlers {
		handler(g)
	}
}

func addUserRoutes(e *echo.Echo, endpoints users.UserEndpointsHandler, middlewares ...echo.MiddlewareFunc) {
	g := e.Group("/users", middlewares...)
	g.GET("/:id", endpoints.GetUser)
	g.GET("", endpoints.GetUsers)
	g.POST("", endpoints.CreateUser)
	g.PUT("/:id", endpoints.UpdateUser)
	g.DELETE("/:id", endpoints.DeleteUser)
}

func addCourseRoutes(e *echo.Echo, endpoints courses.CourseEndpointsHandler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/courses")
	publicGroup.GET("/:id", endpoints.GetCourse)
	publicGroup.GET("", endpoints.GetCourses)
	protectedGroup := e.Group("/courses", middlewares...)
	protectedGroup.POST("", endpoints.CreateCourse)
	protectedGroup.PUT("/:id", endpoints.UpdateCourse)
	protectedGroup.DELETE("/:id", endpoints.DeleteCourse)
}

func addCoursePeriodRoutes(e *echo.Echo, endpoints course_periods.CoursePeriodEndpointsHandler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/courses/:course_id/periods")
	publicGroup.GET("/:id", endpoints.GetCoursePeriod)
	publicGroup.GET("", endpoints.GetCoursePeriods)
	protectedGroup := e.Group("/courses/:course_id/periods", middlewares...)
	protectedGroup.POST("", endpoints.CreateCoursePeriod)
	protectedGroup.PUT("/:id", endpoints.UpdateCoursePeriod)
	protectedGroup.DELETE("/:id", endpoints.DeleteCoursePeriod)

	// Announcement routes
	publicAnnouncementGroup := e.Group("/course-periods/:period_id/announcements")
	publicAnnouncementGroup.GET("/:id", endpoints.GetAnnouncement)
	protectedAnnouncementGroup := e.Group("/course-periods/:period_id/announcements", middlewares...)
	protectedAnnouncementGroup.POST("", endpoints.CreateAnnouncement)
	protectedAnnouncementGroup.PUT("/:id", endpoints.UpdateAnnouncement)
	protectedAnnouncementGroup.DELETE("/:id", endpoints.DeleteAnnouncement)
}

func addGroupsRoutes(e *echo.Echo, endpoints groups.GroupEndpointsHandler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/groups")
	publicGroup.GET("/:id", endpoints.GetGroup)
	publicGroup.GET("", endpoints.GetGroups)
	protectedGroup := e.Group("/groups/requests", middlewares...)
	protectedGroup.POST("", endpoints.CreateGroup)
	protectedGroup.PUT("/:id", endpoints.UpdateGroup)
	protectedGroup.DELETE("/:id", endpoints.DeleteGroup)
}

func addProviderRoutes(e *echo.Echo, endpoints providers.ProviderEndpointsHandler, middlewares ...echo.MiddlewareFunc) {
	group := e.Group("/providers", middlewares...)
	group.GET("/:id", endpoints.GetProvider)
	group.GET("", endpoints.GetProviders)
	group.POST("", endpoints.CreateProvider)
	group.PUT("/:id", endpoints.UpdateProvider)
	group.DELETE("/:id", endpoints.DeleteProvider)
}
