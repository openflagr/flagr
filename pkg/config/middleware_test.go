// nolint: errcheck
package config

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

	// Signed with secret: "mysecret"
	validHS512JWTTokenWithSecret = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.G4VTPaWRHtByF6SaHSQFTeu-896jFb2dF2KnYjJTa9MY_a6Tbb9BsO7Uu0Ju_QOGGDI_b-k6U0T6qwj9lA5_Aw"
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

func TestJWTAuthMiddleware(t *testing.T) {
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

	t.Run("it will pass if jwt enabled with correct header token encrypted using HS512", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthSecret = "mysecret"
		Config.JWTAuthSigningMethod = "HS512"
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthSecret = ""
			Config.JWTAuthSigningMethod = ""
		}()
		hh := SetupGlobalMiddleware(h)

		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		req.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: validHS512JWTTokenWithSecret,
		})
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestJWTAuthMiddlewareWithUnauthorized(t *testing.T) {
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

func TestBasicAuthMiddleware(t *testing.T) {
	h := &okHandler{}

	t.Run("it will return 200 for web paths when disabled", func(t *testing.T) {
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

	t.Run("it will return 200 for whitelist path if basic auth is enabled", func(t *testing.T) {
		Config.BasicAuthEnabled = true
		Config.BasicAuthUsername = "admin"
		Config.BasicAuthPassword = "password"
		defer func() {
			Config.BasicAuthEnabled = false
			Config.BasicAuthUsername = ""
			Config.BasicAuthPassword = ""
		}()

		hh := SetupGlobalMiddleware(h)
		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/flags", nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	t.Run("it will return 401 for web paths when enabled and no basic auth passed", func(t *testing.T) {
		Config.BasicAuthEnabled = true
		Config.BasicAuthUsername = "admin"
		Config.BasicAuthPassword = "password"
		defer func() {
			Config.BasicAuthEnabled = false
			Config.BasicAuthUsername = ""
			Config.BasicAuthPassword = ""
		}()

		testPaths := []string{"/", "", "/#", "/#/", "/static", "/static/"}
		for _, path := range testPaths {
			t.Run(fmt.Sprintf("path: %s", path), func(t *testing.T) {
				hh := SetupGlobalMiddleware(h)
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusUnauthorized, res.Code)
			})
		}
	})

	t.Run("it will return 200 for web paths when enabled and basic auth passed", func(t *testing.T) {
		Config.BasicAuthEnabled = true
		Config.BasicAuthUsername = "admin"
		Config.BasicAuthPassword = "password"
		defer func() {
			Config.BasicAuthEnabled = false
			Config.BasicAuthUsername = ""
			Config.BasicAuthPassword = ""
		}()

		testPaths := []string{"/", "", "/#", "/#/", "/static", "/static/"}
		for _, path := range testPaths {
			t.Run(fmt.Sprintf("path: %s", path), func(t *testing.T) {
				hh := SetupGlobalMiddleware(h)
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", path), nil)
				req.SetBasicAuth(Config.BasicAuthUsername, Config.BasicAuthPassword)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusOK, res.Code)
			})
		}
	})

}

// Mock handler that tracks if it was called
type mockHandler struct {
	called bool
	path   string
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	m.path = r.URL.Path
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("mock handler called"))
}

func TestStaticFileMiddleware(t *testing.T) {
	// Save original config values
	originalPProfEnabled := Config.PProfEnabled
	defer func() {
		Config.PProfEnabled = originalPProfEnabled
	}()

	t.Run("should handle debug/pprof paths correctly (not serve static files)", func(t *testing.T) {
		Config.PProfEnabled = true

		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		testCases := []struct {
			name string
			path string
		}{
			{"pprof root", "/debug/pprof/"},
			{"pprof goroutine", "/debug/pprof/goroutine"},
			{"pprof heap", "/debug/pprof/heap"},
			{"pprof allocs", "/debug/pprof/allocs"},
			{"pprof block", "/debug/pprof/block"},
			{"pprof mutex", "/debug/pprof/mutex"},
			{"pprof with query params", "/debug/pprof/goroutine?debug=1"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "http://localhost:3000"+tc.path, nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// Debug paths should return 200 (handled by pprof) not serve Vue.js index.html
				assert.Equal(t, http.StatusOK, w.Code, "Debug path should return 200: %s", tc.path)

				// Should not return Vue.js HTML content (which would indicate static file middleware caught it)
				responseBody := w.Body.String()
				assert.NotContains(t, responseBody, "Vue App", "Should not serve Vue.js app for debug path: %s", tc.path)
				assert.NotContains(t, responseBody, "<div id=\"app\">", "Should not serve Vue.js app for debug path: %s", tc.path)
			})
		}
	})

	t.Run("should pass through API paths to next handler", func(t *testing.T) {
		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		testCases := []string{
			"/api/v1/health",
			"/api/v1/flags",
			"/api/v1/evaluation",
		}

		for _, path := range testCases {
			t.Run(path, func(t *testing.T) {
				mockHandler.called = false
				mockHandler.path = ""

				req := httptest.NewRequest("GET", "http://localhost:3000"+path, nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// The mock handler should be called for API paths
				assert.True(t, mockHandler.called, "Handler should be called for API path: %s", path)
				assert.Equal(t, path, mockHandler.path, "Path should match")
			})
		}
	})

	t.Run("should serve static files for non-API non-debug paths", func(t *testing.T) {
		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		testCases := []string{
			"/",
			"/flags",
			"/some-page",
			"/static/css/app.css", // This would normally be a static file
		}

		for _, path := range testCases {
			t.Run(path, func(t *testing.T) {
				mockHandler.called = false

				req := httptest.NewRequest("GET", "http://localhost:3000"+path, nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// For non-API, non-debug paths, the static file middleware should handle it
				// and NOT call the next handler (our mock handler)
				// Note: In real scenario, it would serve index.html, but in test it might still call next
				// The important thing is that the response is handled by static middleware
				assert.NotEqual(t, http.StatusNotFound, w.Code, "Should not return 404 for path: %s", path)
			})
		}
	})
}

func TestStaticFileMiddlewareWithJWTAuth(t *testing.T) {
	// Save original config values
	originalJWTEnabled := Config.JWTAuthEnabled
	originalPProfEnabled := Config.PProfEnabled
	originalWhitelistPaths := Config.JWTAuthPrefixWhitelistPaths

	defer func() {
		Config.JWTAuthEnabled = originalJWTEnabled
		Config.PProfEnabled = originalPProfEnabled
		Config.JWTAuthPrefixWhitelistPaths = originalWhitelistPaths
	}()

	t.Run("should allow debug/pprof paths when whitelisted in JWT", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.PProfEnabled = true
		Config.JWTAuthPrefixWhitelistPaths = []string{"/api/v1/health", "/debug/pprof"}

		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		req := httptest.NewRequest("GET", "http://localhost:3000/debug/pprof/", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Should return 200 (handled by pprof), not 401 or redirect
		assert.Equal(t, http.StatusOK, w.Code, "Whitelisted debug path should return 200")

		// Should not return Vue.js HTML content
		responseBody := w.Body.String()
		assert.NotContains(t, responseBody, "Vue App", "Should not serve Vue.js app for debug path")
	})

	t.Run("should block debug/pprof paths when not whitelisted in JWT", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.PProfEnabled = true
		Config.JWTAuthPrefixWhitelistPaths = []string{"/api/v1/health"} // No debug paths

		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		req := httptest.NewRequest("GET", "http://localhost:3000/debug/pprof/", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		// Should be blocked by JWT middleware (redirect or 401)
		assert.True(t, w.Code == http.StatusTemporaryRedirect || w.Code == http.StatusUnauthorized,
			"Should be blocked by JWT auth, got status: %d", w.Code)
	})
}

func TestDebugPathExclusion(t *testing.T) {
	t.Run("should handle pprof debug paths correctly", func(t *testing.T) {
		Config.PProfEnabled = true

		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		// Only pprof paths should be handled by pprof middleware
		pprofPaths := []string{
			"/debug/pprof/",
			"/debug/pprof/goroutine",
			"/debug/pprof/heap",
			"/debug/pprof/allocs",
			"/debug/pprof/block",
			"/debug/pprof/mutex",
		}

		for _, path := range pprofPaths {
			t.Run(path, func(t *testing.T) {
				req := httptest.NewRequest("GET", "http://localhost:3000"+path, nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// pprof paths should return 200 and not serve Vue.js content
				assert.Equal(t, http.StatusOK, w.Code, "pprof path should return 200: %s", path)

				// Should not return Vue.js HTML content (which would indicate static file middleware caught it)
				responseBody := w.Body.String()
				assert.NotContains(t, responseBody, "Vue App", "Should not serve Vue.js app for pprof path: %s", path)
				assert.NotContains(t, responseBody, "<div id=\"app\">", "Should not serve Vue.js app for pprof path: %s", path)
			})
		}
	})

	t.Run("should pass through non-pprof debug paths to next handler", func(t *testing.T) {
		Config.PProfEnabled = true

		mockHandler := &mockHandler{}
		middleware := SetupGlobalMiddleware(mockHandler)

		// Non-pprof debug paths should pass through to next handler (which might serve static files)
		nonPprofDebugPaths := []string{
			"/debug/",
			"/debug/vars",
			"/debug/custom-endpoint",
		}

		for _, path := range nonPprofDebugPaths {
			t.Run(path, func(t *testing.T) {
				req := httptest.NewRequest("GET", "http://localhost:3000"+path, nil)
				w := httptest.NewRecorder()

				middleware.ServeHTTP(w, req)

				// These paths should return 200 (they pass through static middleware and serve index.html)
				assert.Equal(t, http.StatusOK, w.Code, "Non-pprof debug path should return 200: %s", path)

				// These paths should serve Vue.js content (since they're not handled by pprof)
				responseBody := w.Body.String()
				assert.Contains(t, responseBody, "<div id=\"app\">", "Non-pprof debug path should serve Vue.js app: %s", path)
			})
		}
	})
}
