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

	t.Run("it will return 401 for some paths", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthNoTokenStatusCode = http.StatusUnauthorized
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthNoTokenStatusCode = http.StatusTemporaryRedirect
		}()

		testPaths := []string{"/api/v1/flags", "/api/v1/health/..", "/api/v1/admin", "//api/v1/flags", "/..", "/."}
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

	t.Run("current-dir dots do not skip JWT whitelist", func(t *testing.T) {
		Config.JWTAuthEnabled = true
		Config.JWTAuthNoTokenStatusCode = http.StatusUnauthorized
		defer func() {
			Config.JWTAuthEnabled = false
			Config.JWTAuthNoTokenStatusCode = http.StatusTemporaryRedirect
		}()
		hh := SetupGlobalMiddleware(h)

		// "." / "././" collapse in place: still the whitelisted evaluation path.
		for _, p := range []string{"/api/v1/./evaluation", "/api/v1/evaluation/./", "/api/v1/health/././"} {
			t.Run(p+" allowed", func(t *testing.T) {
				res := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "http://localhost:18000"+p, nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusOK, res.Code)
			})
		}
		// Collapsed path is /api/v1/flags, which is not whitelisted.
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://localhost:18000/api/v1/./flags", nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
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

func TestRejectDotDotPath(t *testing.T) {
	h := &okHandler{}
	hh := SetupGlobalMiddleware(h)

	t.Run("it rejects paths containing .. with 401", func(t *testing.T) {
		for _, p := range []string{
			"/api/v1/health/../flags",
			"/api/v1/xx/../flags",
			"/../api/v1/flags",
			"/..",
			"/api/v1/evaluation/../evaluation",
		} {
			t.Run(p, func(t *testing.T) {
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:18000%s", p), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusUnauthorized, res.Code)
			})
		}
	})

	t.Run("it does not reject clean paths", func(t *testing.T) {
		for _, p := range []string{"/api/v1/flags", "/api/v1/health", "/api/v1/evaluation", "/.", "/api/v1/foo..bar", "/api/v1/./flags", "/api/v1/././evaluation"} {
			t.Run(p, func(t *testing.T) {
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:18000%s", p), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusOK, res.Code)
			})
		}
	})

	t.Run("it rejects encoded and backslash parent segments", func(t *testing.T) {
		for _, tc := range []struct {
			path    string
			rawPath string
		}{
			{path: "/api/v1/../flags", rawPath: ""},
			{path: "/api/v1/%2e%2e/flags", rawPath: "/api/v1/%2e%2e/flags"},
			{path: "/api/v1/%252e%252e/flags", rawPath: "/api/v1/%252e%252e/flags"},
			{path: `/api/v1/health\..\flags`, rawPath: ""},
		} {
			t.Run(tc.path, func(t *testing.T) {
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("POST", "http://localhost:18000/x", nil)
				req.URL.Path = tc.path
				req.URL.RawPath = tc.rawPath
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusUnauthorized, res.Code)
			})
		}
	})
}

func TestEvalOnlyDenyMiddleware(t *testing.T) {
	h := &okHandler{}

	setEvalOnly := func(t *testing.T) {
		Config.EvalOnlyMode = true
		t.Cleanup(func() { Config.EvalOnlyMode = false })
	}

	t.Run("it will return 403 for writes under /api/v1/flags in eval-only mode", func(t *testing.T) {
		setEvalOnly(t)
		hh := SetupGlobalMiddleware(h)

		writes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/flags"},
			{"PUT", "/api/v1/flags/1"},
			{"DELETE", "/api/v1/flags/1"},
			{"PUT", "/api/v1/flags/1/enabled"},
			{"POST", "/api/v1/flags/1/variants"},
			{"PUT", "/api/v1/flags/1/segments/2/distributions"},
			{"DELETE", "/api/v1/flags/1/segments/2/constraints/3"},
			{"POST", "/api/v1/flags/1/tags"},
			{"POST", "//api/v1/flags"},
			{"PUT", "//api/v1/flags/1/enabled"},
			{"DELETE", "/api/v1/flags/./1"},
			{"POST", "/api/v1/./flags"},
			{"POST", "/api/v1/././flags"},
			{"POST", "/api/v1/flags/"},
		}
		for _, w := range writes {
			t.Run(fmt.Sprintf("%s %s", w.method, w.path), func(t *testing.T) {
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest(w.method, fmt.Sprintf("http://localhost:18000%s", w.path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusForbidden, res.Code)
				assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
				assert.Contains(t, res.Body.String(), "read-only (eval-only) mode")
			})
		}
	})

	t.Run("it will pass through reads and non-flags paths in eval-only mode", func(t *testing.T) {
		setEvalOnly(t)
		hh := SetupGlobalMiddleware(h)

		passes := []struct {
			method string
			path   string
		}{
			{"GET", "/api/v1/flags"},
			{"GET", "/api/v1/flags/1"},
			{"GET", "/api/v1/export/eval_cache/json"},
			{"GET", "/api/v1/health"},
			{"POST", "/api/v1/evaluation"},
			{"POST", "/api/v1/evaluation/batch"},
			{"OPTIONS", "/api/v1/flags"},
		}
		for _, p := range passes {
			t.Run(fmt.Sprintf("%s %s", p.method, p.path), func(t *testing.T) {
				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest(p.method, fmt.Sprintf("http://localhost:18000%s", p.path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusOK, res.Code)
			})
		}
	})

	t.Run("it will honor WebPrefix when matching the flags path", func(t *testing.T) {
		setEvalOnly(t)
		Config.WebPrefix = "/flagr"
		defer func() { Config.WebPrefix = "" }()
		hh := SetupGlobalMiddleware(h)

		for _, p := range []string{
			"/flagr/api/v1/flags",
			"/flagr//api/v1/flags",
		} {
			res := httptest.NewRecorder()
			res.Body = new(bytes.Buffer)
			req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:18000%s", p), nil)
			hh.ServeHTTP(res, req)
			assert.Equal(t, http.StatusForbidden, res.Code, p)
		}
	})

	t.Run("it will honor a trailing-slash WebPrefix", func(t *testing.T) {
		setEvalOnly(t)

		for _, tc := range []struct {
			prefix string
			path   string
		}{
			{"/flagr/", "/flagr/api/v1/flags"},
			{"/flagr/", "/flagr//api/v1/flags"},
			{"/", "/api/v1/flags"},
			{"/", "//api/v1/flags"},
			{"/flagr//", "/flagr///api/v1/flags"},
		} {
			t.Run(fmt.Sprintf("prefix %q path %s", tc.prefix, tc.path), func(t *testing.T) {
				Config.WebPrefix = tc.prefix
				defer func() { Config.WebPrefix = "" }()
				hh := SetupGlobalMiddleware(h)

				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:18000%s", tc.path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusForbidden, res.Code)
			})
		}
	})

	t.Run("dot-dot paths are rejected with 401 before the flags deny", func(t *testing.T) {
		setEvalOnly(t)

		for _, tc := range []struct {
			prefix string
			path   string
		}{
			{"", "/api/v1/xx/../flags"},
			{"", "/api/v1/health/../flags"},
			{"", "/../api/v1/flags"},
			{"/flagr", "/flagr/api/v1/xx/../flags"},
			{"/a/b", "/a/b/../api/v1/flags"},
			{"/a/b/c", "/a/b/c/../../api/v1/flags"},
		} {
			t.Run(fmt.Sprintf("prefix %q path %s", tc.prefix, tc.path), func(t *testing.T) {
				Config.WebPrefix = tc.prefix
				defer func() { Config.WebPrefix = "" }()
				hh := SetupGlobalMiddleware(h)

				res := httptest.NewRecorder()
				res.Body = new(bytes.Buffer)
				req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:18000%s", tc.path), nil)
				hh.ServeHTTP(res, req)
				assert.Equal(t, http.StatusUnauthorized, res.Code)
			})
		}
	})

	t.Run("it will not block writes when eval-only mode is off", func(t *testing.T) {
		hh := SetupGlobalMiddleware(h)
		res := httptest.NewRecorder()
		res.Body = new(bytes.Buffer)
		req, _ := http.NewRequest("POST", "http://localhost:18000/api/v1/flags", nil)
		hh.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestIsFlagsAPIPath(t *testing.T) {
	t.Parallel()
	d := newEvalOnlyDeny(&okHandler{})
	cases := []struct {
		path string
		want bool
	}{
		{path: "/api/v1/flags", want: true},
		{path: "/api/v1/flags/", want: true},
		{path: "/api/v1/flags/1", want: true},
		{path: "//api/v1/flags", want: true},
		{path: "/api/v1/flags/./1", want: true},
		{path: "api/v1/flags", want: true},
		{path: "/api/v1/evaluation", want: false},
		{path: "/api/v1/health", want: false},
		{path: "/api/v1/export/eval_cache/json", want: false},
		// ".." never reaches this matcher (rejectDotDotPath). HasSafePrefix
		// also refuses them, so they are not classified as flags writes.
		{path: "/api/v1/health/../flags", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, d.isFlagsAPIPath(tc.path))
		})
	}
}
