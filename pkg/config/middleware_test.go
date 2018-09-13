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

const (
	// Signed with secret: ""
	validHS256JWTToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmbGFncl91c2VyIjoiMTIzNDU2Nzg5MCJ9.CLXgNEtwPCqCOtUU-KmqDyO8S2wC_G6PZ0tml8DCuNw"

	// Public Key:
	//-----BEGIN PUBLIC KEY-----
	//MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDdlatRjRjogo3WojgGHFHYLugd
	//UWAY9iR3fy4arWNA1KoS8kVw33cJibXr8bvwUAUparCwlvdbH6dvEOfou0/gCFQs
	//HUfQrSDv+MuSUMAe8jzKE4qW+jK+xQU9a03GUnKHkkle+Q0pX/g6jXZ7r1/xAK5D
	//o2kQ+X5xK9cipRgEKwIDAQAB
	//-----END PUBLIC KEY-----
	validRS256JWTToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TCYt5XsITJX1CxPCT8yAV-TVkIEq_PbChOMqsLfRoPsnsgw5WEuts01mq-pQy7UJiN5mgRxD-WUcX16dUEMGlv50aqzpqh4Qktb3rk-BuQy72IFLOqV0G_zS245-kronKb78cPN25DGlcTwLtjPAYuNzVBAh4vGHSrQyHUdBBPM"

	// Signed with secret: "mysecret"
	validHS256JWTTokenWithSecret = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.drt_po6bHhDOF_FJEHTrK-KD8OGjseJZpHwHIgsnoTM"
)

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
			Value: validHS256JWTToken,
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
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", Config.JWTAuthPrefixWhitelistPaths[0]), nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will pass if jwt enabled with correct header token", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		defer func() { Config.JWTAuthEnabled = false }()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.Header.Add("Authorization", "Bearer "+validHS256JWTToken)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will redirect if jwt enabled with invalid cookie token and valid header token", func(t *testing.T) {
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
		req.Header.Add("Authorization", "Bearer "+validHS256JWTToken)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	})

	t.Run("it will redirect if jwt enabled and a cookie token encrypted with the wrong method", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthSigningMethod = "RS256"
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthSigningMethod = "HS256"
		}()
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

	t.Run("it will pass if jwt enabled with correct header token encrypted using RS256", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthSigningMethod = "RS256"
		Config.JWTAuthSecret = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDdlatRjRjogo3WojgGHFHYLugd
UWAY9iR3fy4arWNA1KoS8kVw33cJibXr8bvwUAUparCwlvdbH6dvEOfou0/gCFQs
HUfQrSDv+MuSUMAe8jzKE4qW+jK+xQU9a03GUnKHkkle+Q0pX/g6jXZ7r1/xAK5D
o2kQ+X5xK9cipRgEKwIDAQAB
-----END PUBLIC KEY-----`
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthSigningMethod = "HS256"
			Config.JWTAuthSecret = ""
		}()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.Header.Add("Authorization", "Bearer "+validRS256JWTToken)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will pass if jwt enabled with valid cookie token with passphrase", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthSecret = "mysecret"
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthSecret = ""
		}()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validHS256JWTTokenWithSecret,
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will pass with a correct HS256 token cookie when signing method is wrong and it defaults to empty string secret", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthSigningMethod = "invalid"
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthSigningMethod = "HS256"
		}()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validHS256JWTToken,
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestAuthMiddlewareWithUnauthorized(t *testing.T) {
	h := &okHandler{}

	t.Run("it will return 401 if no cookie passed", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthNoTokenStatusCode = http.StatusUnauthorized
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthNoTokenStatusCode = http.StatusTemporaryRedirect
		}()

		hh := SetupGlobalMiddleware(h)
		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("it will return 200 if cookie passed", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthNoTokenStatusCode = http.StatusUnauthorized
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthNoTokenStatusCode = http.StatusTemporaryRedirect
		}()

		hh := SetupGlobalMiddleware(h)
		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validHS256JWTToken,
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will return 200 for some paths", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthNoTokenStatusCode = http.StatusUnauthorized
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthNoTokenStatusCode = http.StatusTemporaryRedirect
		}()

		testPaths := []string{"/", "", "/#", "/#/", "/static", "/static/"}
		for _, path := range testPaths {
			t.Run(fmt.Sprintf("path: %s", path), func(t *testing.T) {
				hh := SetupGlobalMiddleware(h)
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusOK, res.Code)
			})
		}
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
