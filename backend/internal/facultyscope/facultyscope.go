// Package facultyscope centralizes the one security-critical decision shared
// by every admin "list requests" endpoint: which faculty's data the caller
// may see. A faculty-scoped admin (role faculty_admin) must never see another
// faculty's requests, even if they pass a different ?faculty= query param —
// the JWT "faculty" claim is the only trustworthy source for them. The query
// param is only ever honored for the DEU-wide admin.
package facultyscope

import (
	"github.com/eaguilar88/deu/internal/entities"
	"github.com/labstack/echo/v4"
)

const facultyContextKey = "faculty" // matches jwt/middleware.go's c.Set("faculty", ...)

// IsGlobalAdmin reports whether the caller's JWT faculty claim is DEU-wide.
func IsGlobalAdmin(c echo.Context) bool {
	return ownFaculty(c) == entities.FacultyDEU
}

func ownFaculty(c echo.Context) entities.Faculty {
	if f, ok := c.Get(facultyContextKey).(string); ok {
		return entities.Faculty(f)
	}
	return ""
}

// Resolve is for endpoints that require a single faculty filter.
// Global admin: requestedFaculty is validated and returned as-is.
// Faculty-scoped admin: requestedFaculty is ignored; their own faculty is
// returned instead, silently overriding whatever they asked for.
func Resolve(c echo.Context, requestedFaculty string) (entities.Faculty, error) {
	if IsGlobalAdmin(c) {
		return entities.FromString(requestedFaculty)
	}
	return ownFaculty(c), nil
}

// ResolveOptional is for endpoints where "no filter" is a valid state.
// Global admin: empty requestedFaculty means "no filter" (""), non-empty is
// validated. Faculty-scoped admin: always their own faculty, never empty.
func ResolveOptional(c echo.Context, requestedFaculty string) (entities.Faculty, error) {
	if IsGlobalAdmin(c) {
		if requestedFaculty == "" {
			return "", nil
		}
		return entities.FromString(requestedFaculty)
	}
	return ownFaculty(c), nil
}
