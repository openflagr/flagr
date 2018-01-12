package config

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupGlobalMiddleware(t *testing.T) {
	var h, hh http.Handler

	hh = SetupGlobalMiddleware(h)
	assert.NotNil(t, hh)

	Config.NewRelicEnabled = true
	hh = SetupGlobalMiddleware(h)
	assert.NotNil(t, hh)
	Config.NewRelicEnabled = false

	Config.JWTAuthEnabled = true
	hh = SetupGlobalMiddleware(h)
	assert.NotNil(t, hh)
	Config.JWTAuthEnabled = false

	Config.PProfEnabled = false
	hh = SetupGlobalMiddleware(h)
	assert.NotNil(t, hh)
	Config.PProfEnabled = true
}
