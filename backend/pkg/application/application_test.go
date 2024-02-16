package application

import (
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestEndpointsConstants(t *testing.T) {
	assert.EqualValues(t, "./docs/openapi/service.yaml", docsSource)
}

func TestApplication_New(t *testing.T) {
	setTestConfigEnvs(t)
	app, err := New(log.NewNopLogger())
	assert.Nil(t, err)
	assert.NotNil(t, app)
}

func TestApplication_Run(t *testing.T) {
	setTestConfigEnvs(t)
	// we need to make it fail since its for test purposes
	assert.NoError(t, os.Setenv("HTTP_SERVE_PORT", "99999999"))
	app, err := New(log.NewNopLogger())
	assert.Nil(t, err)
	assert.NotNil(t, app)

	cn := make(chan error, 1)
	go func() {
		cn <- app.Run()
	}()
	err = <-cn
	assert.Error(t, err)
	assert.EqualError(t, err, "listen tcp: address 99999999: invalid port")
}

func TestApplication_NewErrEnvs(t *testing.T) {
	setTestConfigEnvs(t)
	assert.NoError(t, os.Unsetenv("JWT_SIGNING_KEY"))
	app, err := New(log.NewNopLogger())
	assert.Nil(t, app)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "required environment variable \"JWT_SIGNING_KEY\" is not set")
}

func TestApplication_NewErrInvalidKey(t *testing.T) {
	setTestConfigEnvs(t)
	assert.NoError(t, os.Setenv("JWT_SIGNING_KEY", ""))
	app, err := New(log.NewNopLogger())
	assert.Nil(t, app)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid key length")
}

func TestApplication_buildConfig(t *testing.T) {
	setTestConfigEnvs(t)
	app := application{
		logger: log.NewNopLogger(),
	}
	_, err := app.buildConfig()
	assert.Nil(t, err)
}

func setTestConfigEnvs(t *testing.T) {
	assert.NoError(t, os.Setenv("JWT_SIGNING_KEY", "testKey--testKey"))
	assert.NoError(t, os.Setenv("CFA_PRODUCT_JWT", "testKey--testKey"))
	assert.NoError(t, os.Setenv("CLOUD_CART_JWT", "testKey--testKey"))
	assert.NoError(t, os.Setenv("CFA_VENUE_JWT", "testKey--testKey"))
	assert.NoError(t, os.Setenv("REDUCTION_JWT", "testKey--testKey"))
	assert.NoError(t, os.Setenv("PRODUCT_CATALOG_JWT", "testKey--testKey"))
	assert.NoError(t, os.Setenv("APL_CLOUD_API_JWT_SECRET", "testKey--testKey"))
	assert.NoError(t, os.Setenv("REDUCTION_JWT_EXP_SECONDS", "1000"))
	assert.NoError(t, os.Setenv("PRODUCT_CATALOG_JWT_EXP_SECONDS", "1000"))
}
