package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/eaguilar88/deu/docs"
	"github.com/eaguilar88/deu/internal/activities"
	"github.com/eaguilar88/deu/internal/auth"
	"github.com/eaguilar88/deu/internal/config"
	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/course_periods"
	"github.com/eaguilar88/deu/internal/course_requests"
	"github.com/eaguilar88/deu/internal/courses"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/files"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/group_resource_requests"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/jwt"
	repository "github.com/eaguilar88/deu/internal/postgres_repository"
	"github.com/eaguilar88/deu/internal/provider_requests"
	"github.com/eaguilar88/deu/internal/providers"
	"github.com/eaguilar88/deu/internal/security"
	"github.com/eaguilar88/deu/internal/storage"
	"github.com/eaguilar88/deu/internal/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

//go:embed VERSION
var appVersion string

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
		config.BaseURL,
		logger,
	)
	if err != nil {
		logger.Error("error creating b2 client", zap.Error(err))
		os.Exit(1)
	}

	repository := repository.NewRepository(postgres, config.FilePath, logger)
	authService := auth.NewService(repository, signer, logger)
	authEndpoints := auth.NewHandler(authService)
	mailClient := email.NewMailgunClient(config.Email, logger)
	userSvc := users.NewService(repository, mailClient, bbClient, logger)
	userEndpoints := users.NewHandler(userSvc, logger)

	providerService := providers.NewService(repository, bbClient, mailClient, logger)
	providerEndpoints := providers.NewHandler(providerService, logger)

	courseSvc := courses.NewService(repository, bbClient, logger)
	courseEndpoints := courses.NewHandler(courseSvc, logger)

	cpService := course_periods.NewService(repository, logger)
	cpEndpoints := course_periods.NewHandler(cpService, logger)

	activityService := activities.NewService(repository, bbClient, logger)
	activityEndpoints := activities.NewHandler(activityService, logger)

	groupService := groups.NewService(repository, logger)
	groupEndpoints := groups.NewHandler(groupService, logger)

	groupRequestService := group_requests.NewService(repository, logger)
	groupRequestEndpoints := group_requests.NewHandler(groupRequestService, logger)

	groupResourceRequestService := group_resource_requests.NewService(repository, logger)
	groupResourceRequestEndpoints := group_resource_requests.NewHandler(groupResourceRequestService, logger)

	courseRequestService := course_requests.NewService(repository, logger)
	courseRequestEndpoints := course_requests.NewHandler(courseRequestService, logger)

	providerRequestService := provider_requests.NewService(repository, mailClient, logger)
	providerRequestEndpoints := provider_requests.NewHandler(providerRequestService, logger)

	cycleCloseService := course_cycle_close_requests.NewService(repository, logger)
	cycleCloseEndpoints := course_cycle_close_requests.NewHandler(cycleCloseService, logger)

	e := echo.New()
	e.Validator = security.NewCustomValidator()
	e.HTTPErrorHandler = httperrors.NewHTTPErrorHandler(logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	middlewares := []echo.MiddlewareFunc{
		jwt.JWTMiddleware(signer, logger),
	}

	fileHandler := files.NewHandler(bbClient, logger)

	docs.RegisterDocsRoute(e, logger)
	addHealthRoute(e)
	addVersionRoute(e, strings.TrimSpace(appVersion))
	addFileRoutes(e, fileHandler)
	addAuthRoutes(e, authEndpoints)
	addUserRoutes(e, userEndpoints, middlewares...)
	addProviderRoutes(e, providerEndpoints, middlewares...)
	addCourseRoutes(e, courseEndpoints, middlewares...)
	addCoursePeriodRoutes(e, cpEndpoints, middlewares...)
	addGroupsRoutes(e, groupEndpoints, middlewares...)
	addActivityRoutes(e, activityEndpoints, middlewares...)
	addGroupResourceRequestRoutes(e, groupResourceRequestEndpoints, middlewares...)

	addAdminRoutes(e, middlewares,
		providerEndpoints.RegisterProviderAdminEndpoints,
		groupRequestEndpoints.RegisterGroupRequestAdminEndpoints,
		groupResourceRequestEndpoints.RegisterGroupResourceRequestAdminEndpoints,
		courseRequestEndpoints.RegisterCourseRequestAdminEndpoints,
		providerRequestEndpoints.RegisterProviderRequestAdminEndpoints,
		cycleCloseEndpoints.RegisterAdminEndpoints,
	)

	addCourseCycleCloseRequestRoutes(e, cycleCloseEndpoints, middlewares...)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.HTTPPort)))
}

func addFileRoutes(e *echo.Echo, handler *files.Handler) {
	e.GET("/files/*key", handler.ServeFile)
}

func addHealthRoute(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Ok")
	})
}

func addVersionRoute(e *echo.Echo, version string) {
	e.GET("/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"version": version})
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

func addAuthRoutes(e *echo.Echo, endpoints *auth.Handler) {
	e.POST("/auth/login", endpoints.LoginHandleHTTP)
}

func addAdminRoutes(e *echo.Echo, middlewares []echo.MiddlewareFunc, handlers ...RegisterAdminEndpoints) {
	g := e.Group("/admin", middlewares...)
	for _, handler := range handlers {
		handler(g)
	}
}

func addUserRoutes(e *echo.Echo, endpoints *users.Handler, middlewares ...echo.MiddlewareFunc) {
	public := e.Group("/users")
	public.GET("/:id", endpoints.GetUser)
	public.POST("", endpoints.CreateUser)
	protected := e.Group("/users", middlewares...)
	protected.GET("", endpoints.GetUsers)
	protected.PUT("/:id", endpoints.UpdateUser)
	protected.DELETE("/:id", endpoints.DeleteUser)
}

func addCourseRoutes(e *echo.Echo, endpoints *courses.Handler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/courses")
	publicGroup.GET("/:id", endpoints.GetCourse)
	publicGroup.GET("", endpoints.GetCourses)
	protectedGroup := e.Group("/courses", middlewares...)
	protectedGroup.POST("", endpoints.CreateCourse)
	protectedGroup.PUT("/:id", endpoints.UpdateCourse)
	protectedGroup.DELETE("/:id", endpoints.DeleteCourse)
}

func addCoursePeriodRoutes(e *echo.Echo, endpoints *course_periods.Handler, middlewares ...echo.MiddlewareFunc) {
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

func addGroupResourceRequestRoutes(e *echo.Echo, endpoints *group_resource_requests.Handler, middlewares ...echo.MiddlewareFunc) {
	protected := e.Group("", middlewares...)
	endpoints.RegisterGroupResourceRequestEndpoints(protected)
}

func addActivityRoutes(e *echo.Echo, endpoints *activities.Handler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/activities")
	publicGroup.GET("/:id", endpoints.GetActivity)
	publicGroup.GET("", endpoints.GetActivities)
	protectedGroup := e.Group("/activities", middlewares...)
	protectedGroup.POST("", endpoints.CreateActivity)
	protectedGroup.PUT("/:id", endpoints.UpdateActivity)
	protectedGroup.DELETE("/:id", endpoints.DeleteActivity)
}

func addGroupsRoutes(e *echo.Echo, endpoints *groups.Handler, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/groups")
	publicGroup.GET("/:id", endpoints.GetGroup)
	publicGroup.GET("", endpoints.GetGroups)
	protectedGroup := e.Group("/groups/requests", middlewares...)
	protectedGroup.POST("", endpoints.CreateGroup)
	protectedGroup.PUT("/:id", endpoints.UpdateGroup)
	protectedGroup.DELETE("/:id", endpoints.DeleteGroup)
}

func addCourseCycleCloseRequestRoutes(e *echo.Echo, endpoints *course_cycle_close_requests.Handler, middlewares ...echo.MiddlewareFunc) {
	protected := e.Group("", middlewares...)
	endpoints.RegisterProtectedEndpoints(protected)
}

func addProviderRoutes(e *echo.Echo, endpoints *providers.Handler, middlewares ...echo.MiddlewareFunc) {
	group := e.Group("/providers", middlewares...)
	group.GET("/:id", endpoints.GetProvider)
	group.GET("", endpoints.GetProviders)
	group.POST("", endpoints.CreateProvider)
	group.POST("/documents", endpoints.UploadProviderDocuments)
	group.PUT("/:id", endpoints.UpdateProvider)
	group.DELETE("/:id", endpoints.DeleteProvider)
}
