// nolint: errcheck
package config

import (
	"crypto/subtle"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Allen-Career-Institute/flagr/pkg/config/jwtmiddleware"

	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/gohttp/pprof"
	"github.com/golang-jwt/jwt/v5"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	negroninewrelic "github.com/yadvendar/negroni-newrelic-go-agent"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	IndexFilePath            = "./browser/flagr-ui/dist/index.html"
	StaticFilesDirectoryPath = "./browser/flagr-ui/dist/"
	APIURLPath               = "/api/"
)

// ServerShutdown is a callback function that will be called when
// we tear down the flagr server
func ServerShutdown() {
	if Config.StatsdEnabled && Config.StatsdAPMEnabled {
		tracer.Stop()
	}
}

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	n := negroni.New()

	applyOptionalMiddlewares(n)
	applyStaticFileMiddleware(n)
	applyFinalHandler(n, handler)

	return n
}

// applyOptionalMiddlewares applies all optional middlewares based on configuration
func applyOptionalMiddlewares(n *negroni.Negroni) {
	if Config.MiddlewareGzipEnabled {
		n.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	if Config.MiddlewareVerboseLoggerEnabled {
		n.Use(createVerboseLoggerMiddleware())
	}

	if Config.StatsdEnabled {
		n.Use(&statsdMiddleware{StatsdClient: Global.StatsdClient})
		if Config.StatsdAPMEnabled {
			startAPMTracer()
		}
	}

	if Config.PrometheusEnabled {
		n.Use(&prometheusMiddleware{
			counter:   Global.Prometheus.RequestCounter,
			latencies: Global.Prometheus.RequestHistogram,
		})
	}

	if Config.NewRelicEnabled {
		n.Use(&negroninewrelic.Newrelic{Application: &Global.NewrelicApp})
	}

	if Config.CORSEnabled {
		n.Use(createCORSMiddleware())
	}

	if Config.JWTAuthEnabled {
		n.Use(setupJWTAuthMiddleware())
	}

	if Config.BasicAuthEnabled {
		n.Use(setupBasicAuthMiddleware())
	}
}

// createVerboseLoggerMiddleware creates the verbose logger middleware
func createVerboseLoggerMiddleware() negroni.Handler {
	middleware := negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr")
	for _, u := range Config.MiddlewareVerboseLoggerExcludeURLs {
		middleware.ExcludeURL(u)
	}
	return middleware
}

// startAPMTracer starts the APM tracer for Statsd
func startAPMTracer() {
	tracer.Start(
		tracer.WithAgentAddr(fmt.Sprintf("%s:%s", Config.StatsdHost, Config.StatsdAPMPort)),
		tracer.WithServiceName(Config.StatsdAPMServiceName),
	)
}

// createCORSMiddleware creates the CORS middleware based on the configuration
func createCORSMiddleware() negroni.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   Config.CORSAllowedOrigins,
		AllowedHeaders:   Config.CORSAllowedHeaders,
		ExposedHeaders:   Config.CORSExposedHeaders,
		AllowedMethods:   Config.CORSAllowedMethods,
		AllowCredentials: Config.CORSAllowCredentials,
		MaxAge:           Config.CORSMaxAge,
	})
}

// applyStaticFileMiddleware handles static files and Vue.js routing
func applyStaticFileMiddleware(n *negroni.Negroni) {
	n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if strings.Contains(r.URL.Path, APIURLPath) {
			next(w, r)
			return
		}

		filePath := StaticFilesDirectoryPath + r.URL.Path
		if _, err := os.Stat(filePath); err == nil && filepath.Ext(r.URL.Path) != "" {
			http.ServeFile(w, r, filePath) // Serve the static file directly
			return
		}

		// Serve index.html for Vue.js routing
		http.ServeFile(w, r, IndexFilePath)
	})
}

// applyFinalHandler applies the final handler and sets up the recovery middleware
func applyFinalHandler(n *negroni.Negroni, handler http.Handler) {
	n.Use(setupRecoveryMiddleware())

	if Config.WebPrefix != "" {
		handler = http.StripPrefix(Config.WebPrefix, handler)
	}

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}
}

type recoveryLogger struct{}

func (r *recoveryLogger) Printf(format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

func (r *recoveryLogger) Println(v ...interface{}) {
	logrus.Errorln(v...)
}

func setupRecoveryMiddleware() *negroni.Recovery {
	r := negroni.NewRecovery()
	r.Logger = &recoveryLogger{}
	return r
}

/*
setupJWTAuthMiddleware setup an JWTMiddleware from the ENV config
*/
func setupJWTAuthMiddleware() *jwtAuth {
	var signingMethod jwt.SigningMethod
	var validationKey interface{}
	var errParsingKey error

	switch Config.JWTAuthSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
		validationKey = []byte(Config.JWTAuthSecret)
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
		validationKey = []byte(Config.JWTAuthSecret)
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
		validationKey, errParsingKey = jwt.ParseRSAPublicKeyFromPEM([]byte(Config.JWTAuthSecret))
	default:
		signingMethod = jwt.SigningMethodHS256
		validationKey = []byte("")
	}

	return &jwtAuth{
		PrefixWhitelistPaths: Config.JWTAuthPrefixWhitelistPaths,
		ExactWhitelistPaths:  Config.JWTAuthExactWhitelistPaths,
		JWTMiddleware: jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return validationKey, errParsingKey
			},
			SigningMethod: signingMethod,
			Extractor: jwtmiddleware.FromFirst(
				func(r *http.Request) (string, error) {
					c, err := r.Cookie(Config.JWTAuthCookieTokenName)
					if err != nil {
						return "", nil
					}
					return c.Value, nil
				},
				jwtmiddleware.FromAuthHeader,
			),
			UserProperty: Config.JWTAuthUserProperty,
			Debug:        Config.JWTAuthDebug,
			ErrorHandler: jwtErrorHandler,
		}),
	}
}

func jwtErrorHandler(w http.ResponseWriter, r *http.Request, err string) {
	switch Config.JWTAuthNoTokenStatusCode {
	case http.StatusTemporaryRedirect:
		http.Redirect(w, r, Config.JWTAuthNoTokenRedirectURL, http.StatusTemporaryRedirect)
		return
	default:
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, Config.JWTAuthNoTokenRedirectURL))
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
}

type jwtAuth struct {
	PrefixWhitelistPaths []string
	ExactWhitelistPaths  []string
	JWTMiddleware        *jwtmiddleware.JWTMiddleware
}

func (a *jwtAuth) whitelist(req *http.Request) bool {
	path := req.URL.Path

	// If we set to 401 unauthorized, let the client handles the 401 itself
	if Config.JWTAuthNoTokenStatusCode == http.StatusUnauthorized || Config.JWTAuthNoTokenStatusCode == http.StatusTemporaryRedirect {
		for _, p := range a.ExactWhitelistPaths {
			if p == path {
				return true
			}
		}
	}

	for _, p := range a.PrefixWhitelistPaths {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func (a *jwtAuth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if a.whitelist(req) {
		next(w, req)
		return
	}
	a.JWTMiddleware.HandlerWithNext(w, req, next)
}

/*
setupBasicAuthMiddleware setup an BasicMiddleware from the ENV config
*/
func setupBasicAuthMiddleware() *basicAuth {
	return &basicAuth{
		Username:             []byte(Config.BasicAuthUsername),
		Password:             []byte(Config.BasicAuthPassword),
		PrefixWhitelistPaths: Config.BasicAuthPrefixWhitelistPaths,
		ExactWhitelistPaths:  Config.BasicAuthExactWhitelistPaths,
	}
}

type basicAuth struct {
	Username             []byte
	Password             []byte
	PrefixWhitelistPaths []string
	ExactWhitelistPaths  []string
}

func (a *basicAuth) whitelist(req *http.Request) bool {
	path := req.URL.Path

	for _, p := range a.ExactWhitelistPaths {
		if p == path {
			return true
		}
	}

	for _, p := range a.PrefixWhitelistPaths {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func (a *basicAuth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if a.whitelist(req) {
		next(w, req)
		return
	}

	username, password, ok := req.BasicAuth()
	if !ok || subtle.ConstantTimeCompare(a.Username, []byte(username)) != 1 || subtle.ConstantTimeCompare(a.Password, []byte(password)) != 1 {
		w.Header().Set("WWW-Authenticate", `Basic realm="you shall not pass"`)
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}

	next(w, req)
}

type statsdMiddleware struct {
	StatsdClient *statsd.Client
}

func (s *statsdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func(start time.Time) {
		response := w.(negroni.ResponseWriter)
		status := strconv.Itoa(response.Status())
		duration := float64(time.Since(start)) / float64(time.Millisecond)
		tags := []string{
			"status:" + status,
			"path:" + r.RequestURI,
			"method:" + r.Method,
		}

		s.StatsdClient.Incr("http.requests.count", tags, 1)
		s.StatsdClient.TimeInMilliseconds("http.requests.duration", duration, tags, 1)
	}(time.Now())

	next(w, r)
}

type prometheusMiddleware struct {
	counter   *prometheus.CounterVec
	latencies *prometheus.HistogramVec
}

func (p *prometheusMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.EscapedPath() == Global.Prometheus.ScrapePath {
		handler := promhttp.Handler()
		handler.ServeHTTP(w, r)
	} else {
		defer func(start time.Time) {
			response := w.(negroni.ResponseWriter)
			status := strconv.Itoa(response.Status())
			duration := float64(time.Since(start)) / float64(time.Second)

			p.counter.WithLabelValues(status, r.RequestURI, r.Method).Inc()
			if p.latencies != nil {
				p.latencies.WithLabelValues(status, r.RequestURI, r.Method).Observe(duration)
			}
		}(time.Now())
		next(w, r)
	}
}
