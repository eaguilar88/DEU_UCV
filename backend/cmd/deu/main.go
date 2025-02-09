package main

import (
	"fmt"
	"net/http"
	"os"

	"database/sql"

	"github.com/eaguilar88/deu/docs"
	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/config"
	"github.com/eaguilar88/deu/pkg/courses"
	"github.com/eaguilar88/deu/pkg/endorsements"
	errs "github.com/eaguilar88/deu/pkg/errors"
	"github.com/eaguilar88/deu/pkg/jwt"
	"github.com/eaguilar88/deu/pkg/repository"
	"github.com/eaguilar88/deu/pkg/transport"
	"github.com/eaguilar88/deu/pkg/users"
	kitJWT "github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/endpoint"
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
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

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

	signer := jwt.NewJWTSigner(config.JWTEncryptionKey, config.TTL, &logger)

	r := mux.NewRouter()

	repository := repository.NewRepository(postgres, config.FilePath, logger)
	authService := auth.NewAuthService(repository, signer, logger)
	endpointMiddlewares := []endpoint.Middleware{
		jwt.JWTMiddleware(signer, logger),
	}
	authEndpoints := auth.MakeEndpoints(authService, logger, nil)

	addDocsRoute(r, docsSource, logger)
	userSvc := users.NewUsersService(repository, logger)
	userEndpoints := users.MakeEndpoints(userSvc, logger, endpointMiddlewares)

	endorsementSvc := endorsements.NewEndorsementsService(repository, logger)
	endorsementEndpoints := endorsements.MakeEndpoints(endorsementSvc, logger, endpointMiddlewares)

	courseSvc := courses.NewCoursesService(repository, logger)
	courseEndpoints := courses.MakeEndpoints(courseSvc, logger, endpointMiddlewares)

	commonHTTPOptions := []kitHTTP.ServerOption{
		kitHTTP.ServerBefore(kitJWT.HTTPToContext()),
		kitHTTP.ServerErrorEncoder(errs.MakeHTTPErrorEncoder(logger)),
	}
	addAuthRoutes(r, authEndpoints, commonHTTPOptions)
	addUserRoutes(r, userEndpoints, commonHTTPOptions)
	addEndorsementRoutes(r, endorsementEndpoints, commonHTTPOptions)
	addCourseRoutes(r, courseEndpoints, commonHTTPOptions)

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

func addAuthRoutes(r *mux.Router, endpoints auth.Endpoints, options []kitHTTP.ServerOption) {
	path := fmt.Sprintf(transport.PathAuth, "login")
	r.Methods(http.MethodPost).Path(path).Handler(auth.LoginHandleHTTP(endpoints.Login, options))
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
	createUserHandler := users.CreateUserHandleHTTP(endpoints.CreateUser, options)
	r.Methods(http.MethodPost).Path(transport.PathUsers).Handler(createUserHandler)

	//Update User Endpoint
	updateUserHandler := users.UpdateUserHandleHTTP(endpoints.UpdateUser, options)
	path = fmt.Sprintf(transport.FormatUsers, transport.ParamUserID)
	r.Methods(http.MethodPut).Path(path).Handler(updateUserHandler)

	//Delete User Endpoint
	deleteUserHandler := users.DeleteUserHandleHTTP(endpoints.DeleteUser, options)
	path = fmt.Sprintf(transport.FormatUsers, transport.ParamUserID)
	r.Methods(http.MethodDelete).Path(path).Handler(deleteUserHandler)
}

func addEndorsementRoutes(r *mux.Router, endpoints endorsements.Endpoints, options []kitHTTP.ServerOption) {
	//Get Endorsement Endpoint
	getEndorsementHandler := transport.GetEndorsementHandleHTTP(endpoints.GetEndorsement, options)
	path := fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID)
	r.Methods(http.MethodGet).Path(path).Handler(getEndorsementHandler)

	//Get Endorsements Endpoint
	getEndorsementsHandler := transport.GetEndorsementsHandleHTTP(endpoints.GetEndorsements, options)
	r.Methods(http.MethodGet).Path(transport.PathEndorsements).Handler(getEndorsementsHandler)

	//Create Endorsement Endpoint
	createEndorsementHandler := transport.CreateEndorsementHandleHTTP(endpoints.CreateEndorsement, options)
	r.Methods(http.MethodPost).Path(transport.PathEndorsements).Handler(createEndorsementHandler)

	//Update Endorsement Endpoint
	updateEndorsementHandler := transport.UpdateEndorsementHandleHTTP(endpoints.UpdateEndorsement, options)
	path = fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID)
	r.Methods(http.MethodPut).Path(path).Handler(updateEndorsementHandler)

	//Delete Endorsement Endpoint
	deleteEndorsementHandler := transport.DeleteEndorsementHandleHTTP(endpoints.DeleteEndorsement, options)
	path = fmt.Sprintf(transport.FormatEndorsements, transport.ParamEndorsementID)
	r.Methods(http.MethodDelete).Path(path).Handler(deleteEndorsementHandler)
}

func addCourseRoutes(r *mux.Router, endpoints courses.Endpoints, options []kitHTTP.ServerOption) {
	//Get Course Endpoint
	getCourseHandler := courses.GetCourseHandleHTTP(endpoints.GetCourse, options)
	path := fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID)
	r.Methods(http.MethodGet).Path(path).Handler(getCourseHandler)

	//Get Courses Endpoint
	getCoursesHandler := courses.GetCoursesHandleHTTP(endpoints.GetCourses, options)
	r.Methods(http.MethodGet).Path(transport.PathCourses).Handler(getCoursesHandler)

	//Create Course Endpoint
	createCourseHandler := courses.CreateCourseHandleHTTP(endpoints.CreateCourse, options)
	r.Methods(http.MethodPost).Path(transport.PathCourses).Handler(createCourseHandler)

	//Update Course Endpoint
	updateCourseHandler := courses.UpdateCourseHandleHTTP(endpoints.UpdateCourse, options)
	path = fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID)
	r.Methods(http.MethodPut).Path(path).Handler(updateCourseHandler)

	//Delete Course Endpoint
	deleteCourseHandler := courses.DeleteCourseHandleHTTP(endpoints.DeleteCourse, options)
	path = fmt.Sprintf(transport.FormatCourses, transport.ParamCourseID)
	r.Methods(http.MethodDelete).Path(path).Handler(deleteCourseHandler)
}
