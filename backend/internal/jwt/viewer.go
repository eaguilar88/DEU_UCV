package jwt

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
)

// ViewerFromContext builds the caller's identity from the values the auth middlewares
// store in the context. Without a valid token it returns the zero (anonymous) Viewer.
func ViewerFromContext(c echo.Context) entities.Viewer {
	var viewer entities.Viewer
	if userID, ok := c.Get("userID").(string); ok {
		viewer.UserID = userID
	}
	if roles, ok := c.Get("roles").([]string); ok {
		viewer.Roles = roles
	}
	if faculty, ok := c.Get("faculty").(string); ok {
		viewer.Faculty = entities.Faculty(faculty)
	}
	return viewer
}
