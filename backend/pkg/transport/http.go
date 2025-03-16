package transport

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/eaguilar88/deu/pkg/entities"
	"github.com/labstack/echo/v4"
)

func HealthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Ok")
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
