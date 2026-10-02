package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eaguilar88/deu/docs"
	"github.com/eaguilar88/deu/internal/activities"
	"github.com/eaguilar88/deu/internal/auth"
	"github.com/eaguilar88/deu/internal/certificates"
	"github.com/eaguilar88/deu/internal/config"
	"github.com/eaguilar88/deu/internal/course_cycle_close_requests"
	"github.com/eaguilar88/deu/internal/course_periods"
	"github.com/eaguilar88/deu/internal/course_requests"
	"github.com/eaguilar88/deu/internal/courses"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/files"
	"github.com/eaguilar88/deu/internal/gotenberg"
	"github.com/eaguilar88/deu/internal/group_analytics"
	"github.com/eaguilar88/deu/internal/group_dashboards"
	"github.com/eaguilar88/deu/internal/group_requests"
	"github.com/eaguilar88/deu/internal/group_resource_requests"
	"github.com/eaguilar88/deu/internal/groups"
	"github.com/eaguilar88/deu/internal/httperrors"
	"github.com/eaguilar88/deu/internal/jobs"
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

// shutdownTimeout bounds how long in-flight HTTP requests and the running job get to finish.
const shutdownTimeout = 15 * time.Second

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
	mailClient, err := email.NewSMTPClient(config.Email, logger)
	if err != nil {
		logger.Error("error creating smtp client", zap.Error(err))
		os.Exit(1)
	}
	userSvc := users.NewService(repository, mailClient, bbClient, logger)
	userEndpoints := users.NewHandler(userSvc, logger)

	providerService := providers.NewService(repository, bbClient, mailClient, logger)
	providerEndpoints := providers.NewHandler(providerService, logger)

	courseSvc := courses.NewService(repository, bbClient, mailClient, logger)
	courseEndpoints := courses.NewHandler(courseSvc, logger)

	cpService := course_periods.NewService(repository, logger)
	cpEndpoints := course_periods.NewHandler(cpService, logger)

	activityService := activities.NewService(repository, bbClient, logger)
	activityEndpoints := activities.NewHandler(activityService, logger)

	groupService := groups.NewService(repository, bbClient, logger)
	groupEndpoints := groups.NewHandler(groupService, logger)

	groupRequestService := group_requests.NewService(repository, mailClient, logger)
	groupRequestEndpoints := group_requests.NewHandler(groupRequestService, logger)

	groupResourceRequestService := group_resource_requests.NewService(repository, mailClient, logger)
	groupResourceRequestEndpoints := group_resource_requests.NewHandler(groupResourceRequestService, logger)

	courseRequestService := course_requests.NewService(repository, bbClient, mailClient, logger)
	courseRequestEndpoints := course_requests.NewHandler(courseRequestService, logger)

	providerRequestService := provider_requests.NewService(repository, mailClient, logger)
	providerRequestEndpoints := provider_requests.NewHandler(providerRequestService, logger)

	cycleCloseService := course_cycle_close_requests.NewService(repository, bbClient, mailClient, logger)
	cycleCloseEndpoints := course_cycle_close_requests.NewHandler(cycleCloseService, logger)

	pdfRenderer := gotenberg.NewClient(config.GotenbergURL)
	certificatesService := certificates.NewService(repository, bbClient, mailClient, pdfRenderer, config.PublicBaseURL, logger)
	certificatesEndpoints := certificates.NewHandler(certificatesService, logger)

	worker := jobs.NewWorker(repository, logger)
	worker.Register(entities.JobKindCourseCycleCertificates, certificatesService.GenerateForCloseRequest)

	dashboardSvc := group_dashboards.NewService(repository, logger)
	dashboardEndpoints := group_dashboards.NewHandler(dashboardSvc, logger)

	analyticsSvc := group_analytics.NewService(repository, logger)
	analyticsEndpoints := group_analytics.NewHandler(analyticsSvc, logger)

	e := echo.New()
	e.Validator = security.NewCustomValidator()
	e.HTTPErrorHandler = httperrors.NewHTTPErrorHandler(logger)
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: config.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet, http.MethodHead, http.MethodPost,
			http.MethodPut, http.MethodPatch, http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization,
		},
		AllowCredentials: false, // auth uses the Authorization header, not cookies
	}))
	middlewares := []echo.MiddlewareFunc{
		jwt.JWTMiddleware(signer, logger),
	}
	// Public reads whose response depends on who is asking.
	optionalAuth := jwt.OptionalJWTMiddleware(signer, logger)

	filesSvc := files.NewService(repository, bbClient, logger)
	fileHandler := files.NewHandler(filesSvc, logger)

	docs.RegisterDocsRoute(e, logger)
	addHealthRoute(e)
	addVersionRoute(e, strings.TrimSpace(appVersion))
	addFileRoutes(e, fileHandler)
	addAuthRoutes(e, authEndpoints)
	addUserRoutes(e, userEndpoints, optionalAuth, middlewares...)
	addProviderRoutes(e, providerEndpoints, middlewares...)
	addCourseRoutes(e, courseEndpoints, optionalAuth, middlewares...)
	addCourseRequestRoutes(e, courseRequestEndpoints, middlewares...)
	addCoursePeriodRoutes(e, cpEndpoints, middlewares...)
	addGroupsRoutes(e, groupEndpoints, optionalAuth, middlewares...)
	addActivityRoutes(e, activityEndpoints, optionalAuth, middlewares...)
	addGroupResourceRequestRoutes(e, groupResourceRequestEndpoints, middlewares...)
	addGroupDashboardRoutes(e, dashboardEndpoints, middlewares...)

	// group_admin is the role granted to group-extension admins (see group_requests).
	analyticsMiddlewares := append(append([]echo.MiddlewareFunc{}, middlewares...),
		jwt.RequireRoles("root", "deu_admin", "faculty_admin", "group_admin"),
	)
	addGroupAnalyticsRoutes(e, analyticsEndpoints, analyticsMiddlewares...)

	adminMiddlewares := append(append([]echo.MiddlewareFunc{}, middlewares...),
		jwt.RequireRoles("root", "deu_admin", "faculty_admin"),
	)

	addAdminRoutes(e, adminMiddlewares,
		providerEndpoints.RegisterProviderAdminEndpoints,
		groupRequestEndpoints.RegisterGroupRequestAdminEndpoints,
		groupResourceRequestEndpoints.RegisterGroupResourceRequestAdminEndpoints,
		courseRequestEndpoints.RegisterCourseRequestAdminEndpoints,
		providerRequestEndpoints.RegisterProviderRequestAdminEndpoints,
		cycleCloseEndpoints.RegisterAdminEndpoints,
		certificatesEndpoints.RegisterAdminEndpoints,
		activityEndpoints.RegisterActivityAdminEndpoints,
		dashboardEndpoints.RegisterDashboardAdminEndpoints,
	)

	addCourseCycleCloseRequestRoutes(e, cycleCloseEndpoints, middlewares...)
	certificatesEndpoints.RegisterPublicEndpoints(e)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Run(ctx)
	}()

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", config.HTTPPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error("error shutting down http server", zap.Error(err))
	}

	workerDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workerDone)
	}()
	select {
	case <-workerDone:
	case <-shutdownCtx.Done():
		logger.Warn("jobs worker did not stop in time")
	}
}

func addFileRoutes(e *echo.Echo, handler *files.Handler) {
	// Echo names a wildcard segment "*" whatever follows it, so the key is read with
	// c.Param(files.KeyParam); "/files/*key" would leave c.Param("key") empty.
	e.GET("/files/*", handler.ServeFile)
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

// addUserRoutes serves GET /users/:id with optional auth: a public profile for everyone,
// the full user for the user themself and admins.
func addUserRoutes(e *echo.Echo, endpoints *users.Handler, optionalAuth echo.MiddlewareFunc, middlewares ...echo.MiddlewareFunc) {
	public := e.Group("/users")
	public.GET("/:id", endpoints.GetUser, optionalAuth)
	public.POST("", endpoints.CreateUser)
	protected := e.Group("/users", middlewares...)
	protected.GET("", endpoints.GetUsers)
	protected.PUT("/:id", endpoints.UpdateUser)
	protected.DELETE("/:id", endpoints.DeleteUser)
}

func addCourseRequestRoutes(e *echo.Echo, endpoints *course_requests.Handler, middlewares ...echo.MiddlewareFunc) {
	protectedGroup := e.Group("", middlewares...)
	endpoints.RegisterCourseRequestEndpoints(protectedGroup)
}

// addCourseRoutes serves GET /courses with optional auth: anonymous and regular callers only
// see open or closed courses; admins and a provider listing their own courses see them all.
// GET /courses/:id also takes optional auth, so a provider and its reviewers can open a course
// that is not approved yet.
func addCourseRoutes(e *echo.Echo, endpoints *courses.Handler, optionalAuth echo.MiddlewareFunc, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/courses")
	publicGroup.GET("/public", endpoints.GetPublicCourses)
	publicGroup.GET("/:id", endpoints.GetCourse, optionalAuth)
	publicGroup.GET("", endpoints.GetCourses, optionalAuth)
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

// addActivityRoutes serves GET /activities/:id with optional auth: the participant list is
// only included for viewers who may manage the activity's group.
func addActivityRoutes(e *echo.Echo, endpoints *activities.Handler, optionalAuth echo.MiddlewareFunc, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/activities")
	publicGroup.GET("/:id", endpoints.GetActivity, optionalAuth)
	publicGroup.GET("", endpoints.GetActivities)
	protectedGroup := e.Group("/activities", middlewares...)
	protectedGroup.POST("", endpoints.CreateActivity)
	protectedGroup.PUT("/:id", endpoints.UpdateActivity)
	protectedGroup.PATCH("/feature", endpoints.ToggleFeature)
	protectedGroup.DELETE("/:id", endpoints.DeleteActivity)
	protectedGroup.GET("/group-summary/:groupId", endpoints.GetGroupDashboardSummary)
}

// addGroupsRoutes registers the group reads as public routes with optional auth: the
// response is the public view unless the token's user may see the group's private data.
func addGroupsRoutes(e *echo.Echo, endpoints *groups.Handler, optionalAuth echo.MiddlewareFunc, middlewares ...echo.MiddlewareFunc) {
	publicGroup := e.Group("/groups")
	publicGroup.GET("/:id", endpoints.GetGroup, optionalAuth)
	publicGroup.GET("", endpoints.GetGroups, optionalAuth)
	protectedGroup := e.Group("/groups/requests", middlewares...)
	protectedGroup.POST("", endpoints.CreateGroup)
	protectedGroup.PUT("/:id", endpoints.UpdateGroup)
	protectedGroup.DELETE("/:id", endpoints.DeleteGroup)
}

func addGroupDashboardRoutes(e *echo.Echo, endpoints *group_dashboards.Handler, middlewares ...echo.MiddlewareFunc) {
	protected := e.Group("", middlewares...)
	endpoints.RegisterDashboardProtectedEndpoints(protected)
}

func addGroupAnalyticsRoutes(e *echo.Echo, endpoints *group_analytics.Handler, middlewares ...echo.MiddlewareFunc) {
	protected := e.Group("", middlewares...)
	endpoints.RegisterAnalyticsEndpoints(protected)
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
	group.PUT("/:id", endpoints.UpdateProvider)
	group.DELETE("/:id", endpoints.DeleteProvider)
}
