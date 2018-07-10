package config

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

type okHandler struct{}

func (o *okHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}

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

func TestAuthMiddleware(t *testing.T) {
	h := &okHandler{}

	t.Run("it will redirect if jwt enabled but no cookie passed", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		defer func() { Config.JWTAuthEnabled = false }()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)

		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	})

	t.Run("it will redirect if jwt enabled with wrong cookie passed", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		defer func() { Config.JWTAuthEnabled = false }()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: "invalid_jwt",
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	})

	t.Run("it will pass if jwt enabled with correct cookie passed", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		defer func() { Config.JWTAuthEnabled = false }()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmbGFncl91c2VyIjoiMTIzNDU2Nzg5MCJ9.CLXgNEtwPCqCOtUU-KmqDyO8S2wC_G6PZ0tml8DCuNw", // {"flagr_user": "1234567890"}
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will pass if jwt enabled but with whitelisted path", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		defer func() { Config.JWTAuthEnabled = false }()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", Config.JWTAuthWhitelistPaths[0]), nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestStatsMiddleware(t *testing.T) {
	t.Run("it will setup statsd if statsd is enabled", func(t *testing.T) {
		Config.StatsdEnabled = true
		defer func() { Config.StatsdEnabled = false }()
		hh := SetupGlobalMiddleware(nil)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)

		incrCalled := false
		defer monkey.PatchInstanceMethod(
			reflect.TypeOf(Global.StatsdClient),
			"Incr",
			func(_ *statsd.Client, _ string, _ []string, _ float64) error {
				incrCalled = true
				return nil
			},
		).Unpatch()

		hh.ServeHTTP(res, req)
		assert.True(t, incrCalled)
	})
}
