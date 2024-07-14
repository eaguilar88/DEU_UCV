package transport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/eaguilar88/deu/pkg/auth"
	"github.com/eaguilar88/deu/pkg/entities"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func LoginHandler(ctx context.Context, service *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Login Completed")
	}
}

func NewQueryScopeFromURL(url *url.URL) (entities.PageScope, error) {
	scope := entities.PageScope{}
	vars := url.Query()
	if err := scope.GetPageFromVars(vars.Get(PageParam)); err != nil {
		return scope, fmt.Errorf("error getting page from query string: %v", err)
	}

	if err := scope.GetPerPageFromVars(vars.Get(PerPageParam)); err != nil {
		return scope, fmt.Errorf("error getting per page from query string: %v", err)
	}

	return scope, nil
}

type swaggerVariables struct {
	Version     string
	Name        string
	GitCommitID string
}

// AddDocsEndpoint provides docs endpoint which is used for api documentation
// func AddDocsEndpoint(rtr router.Router, logger log.Logger, docsSource string, middlewares ...router.Middleware) {
// 	rtr.Handle(http.MethodGet, docsURL, func(rw http.ResponseWriter, r *http.Request) {

// 		variables := swaggerVariables{
// 			AppVersion:      constants.AppVersion,
// 			AppName:         constants.AppName,
// 			AppCreationTime: constants.AppCreationTime,
// 			GitBranchName:   constants.GitBranchName,
// 			GitCommitID:     constants.GitCommitID,
// 			GitCommitTime:   constants.GitCommitTime,
// 			GitTags:         constants.GitTags,
// 			BaseURL:         fmt.Sprintf("https://%s", r.Host),
// 		}

// 		if strings.Contains(r.Header.Get("Accept"), "text/yaml") {
// 			rw.Header().Set("Content-Type", docsContentTypeHeader)
// 			rw.Header().Set("Access-Control-Allow-Origin", "*")

// 			tmpl, err := template.ParseFiles(docsSource)
// 			if err != nil {
// 				errors.Handle(logger, err, "failed reading docs")
// 				rw.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}

// 			err = tmpl.Execute(rw, variables)
// 			if err != nil {
// 				errors.Handle(logger, err, "failed writing the data to HTTP response")
// 				rw.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}

// 			rw.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		tmpl, err := template.ParseFS(swaggerFS, "swagger/*.html")
// 		rw.Header().Set("Content-Type", docsHttpTypeHeader)

// 		if err != nil {
// 			errors.Handle(logger, err, "failed reading from index file")
// 			rw.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		err = tmpl.ExecuteTemplate(rw, "index.html", variables)
// 		if err != nil {
// 			errors.Handle(logger, err, "failed writing the data to HTTP response")
// 			rw.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		rw.WriteHeader(http.StatusOK)
// 	}, middlewares...)

// 	rtr.HandlePrefix(http.MethodGet, swaggerURL, func(rw http.ResponseWriter, r *http.Request) {
// 		http.StripPrefix(docsURL, http.FileServer(http.FS(swaggerFS))).ServeHTTP(rw, r)
// 	}, middlewares...)

// }
