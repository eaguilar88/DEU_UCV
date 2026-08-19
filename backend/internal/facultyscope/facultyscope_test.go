package facultyscope_test

import (
	"net/http/httptest"
	"testing"

	"github.com/eaguilar88/deu/internal/entities"
	"github.com/eaguilar88/deu/internal/facultyscope"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newContext(faculty string) echo.Context {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.Set("faculty", faculty)
	return c
}

func TestIsGlobalAdmin(t *testing.T) {
	assert.True(t, facultyscope.IsGlobalAdmin(newContext("DEU")))
	assert.False(t, facultyscope.IsGlobalAdmin(newContext("Ciencias")))
	assert.False(t, facultyscope.IsGlobalAdmin(newContext("")))
}

func TestResolve(t *testing.T) {
	t.Run("global admin gets validated query param", func(t *testing.T) {
		c := newContext("DEU")
		faculty, err := facultyscope.Resolve(c, "Ciencias")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)
	})

	t.Run("global admin invalid faculty errors", func(t *testing.T) {
		c := newContext("DEU")
		_, err := facultyscope.Resolve(c, "NotARealFaculty")
		assert.Error(t, err)
	})

	t.Run("faculty admin is silently overridden to their own faculty", func(t *testing.T) {
		c := newContext("Ciencias")
		faculty, err := facultyscope.Resolve(c, "Medicina")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)
	})

	t.Run("faculty admin with no query param still gets their own faculty", func(t *testing.T) {
		c := newContext("Ciencias")
		faculty, err := facultyscope.Resolve(c, "")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)
	})
}

func TestResolveOptional(t *testing.T) {
	t.Run("global admin with empty query param means no filter", func(t *testing.T) {
		c := newContext("DEU")
		faculty, err := facultyscope.ResolveOptional(c, "")
		require.NoError(t, err)
		assert.Equal(t, entities.Faculty(""), faculty)
	})

	t.Run("global admin with query param filters", func(t *testing.T) {
		c := newContext("DEU")
		faculty, err := facultyscope.ResolveOptional(c, "Ciencias")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)
	})

	t.Run("global admin invalid faculty errors", func(t *testing.T) {
		c := newContext("DEU")
		_, err := facultyscope.ResolveOptional(c, "NotARealFaculty")
		assert.Error(t, err)
	})

	t.Run("faculty admin is always pinned to their own faculty, never empty", func(t *testing.T) {
		c := newContext("Ciencias")
		faculty, err := facultyscope.ResolveOptional(c, "Medicina")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)

		faculty, err = facultyscope.ResolveOptional(c, "")
		require.NoError(t, err)
		assert.Equal(t, entities.FacultyCiencias, faculty)
	})
}
