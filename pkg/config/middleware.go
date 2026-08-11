package config

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gohttp/pprof"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/openflagr/flagr/pkg/util"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	negroninewrelic "github.com/yadvendar/negroni-newrelic-go-agent"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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

	if Config.MiddlewareGzipEnabled {
		n.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	if Config.MiddlewareVerboseLoggerEnabled {
		middleware := negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr")

		for _, u := range Config.MiddlewareVerboseLoggerExcludeURLs {
			middleware.ExcludeURL(u)
		}

		n.Use(middleware)
	}

	if Config.StatsdEnabled {
		n.Use(&statsdMiddleware{StatsdClient: Global.StatsdClient})

		if Config.StatsdAPMEnabled {
			tracer.Start(
				tracer.WithAgentAddr(fmt.Sprintf("%s:%s", Config.StatsdHost, Config.StatsdAPMPort)),
				tracer.WithServiceName(Config.StatsdAPMServiceName),
			)
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
		n.Use(cors.New(cors.Options{
			AllowedOrigins:   Config.CORSAllowedOrigins,
			AllowedHeaders:   Config.CORSAllowedHeaders,
			ExposedHeaders:   Config.CORSExposedHeaders,
			AllowedMethods:   Config.CORSAllowedMethods,
			AllowCredentials: Config.CORSAllowCredentials,
			MaxAge:           Config.CORSMaxAge,
		}))
	}

	setupAuthMiddleware(n)

	if Config.UIEnabled {
		n.Use(&negroni.Static{
			Dir:       http.Dir("./browser/flagr-ui/dist/"),
			Prefix:    Config.WebPrefix,
			IndexFile: "index.html",
		})
	}

	n.Use(setupRecoveryMiddleware())

	if Config.WebPrefix != "" {
		handler = http.StripPrefix(Config.WebPrefix, handler)
	}

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}

	return n
}

type recoveryLogger struct{}

func (r *recoveryLogger) Printf(format string, v ...any) {
	logrus.Errorf(format, v...)
}

func (r *recoveryLogger) Println(v ...any) {
	logrus.Errorln(v...)
}

func setupRecoveryMiddleware() *negroni.Recovery {
	r := negroni.NewRecovery()
	r.Logger = &recoveryLogger{}
	return r
}

// setupAuthMiddleware wires up JWT auth, the optional JWT group-claim check,
// and basic auth, in that order, based on the ENV config.
func setupAuthMiddleware(n *negroni.Negroni) {
	if Config.JWTAuthEnabled {
		n.Use(setupJWTAuthMiddleware())
	}

	if Config.JWTAuthRequireGroupClaim != "" {
		n.Use(setupJWTRequireGroupClaimMiddleware())
	}

	if Config.BasicAuthEnabled {
		n.Use(setupBasicAuthMiddleware())
	}
}

/*
setupJWTAuthMiddleware setup an JWTMiddleware from the ENV config
*/
func setupJWTAuthMiddleware() *jwtAuth {
	var signingMethod jwt.SigningMethod
	var validationKey any
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
			ValidationKeyGetter: func(token *jwt.Token) (any, error) {
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

	if Config.WebPrefix != "" {
		path = strings.TrimPrefix(path, Config.WebPrefix)
	}

	// If we set to 401 unauthorized, let the client handles the 401 itself
	if Config.JWTAuthNoTokenStatusCode == http.StatusUnauthorized {
		if slices.Contains(a.ExactWhitelistPaths, path) {
			return true
		}
	}

	for _, p := range a.PrefixWhitelistPaths {
		if p != "" && util.HasSafePrefix(path, p) {
			return true
		}
	}
	return false
}

type whiteListedCtxKey struct{}

func (a *jwtAuth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if a.whitelist(req) {
		req = req.WithContext(context.WithValue(req.Context(), whiteListedCtxKey{}, true))
		next(w, req)
		return
	}
	a.JWTMiddleware.HandlerWithNext(w, req, next)
}

/*
setupJWTRequireGroupClaimMiddleware sets up a middleware that, when
FLAGR_JWT_AUTH_REQUIRE_GROUP_CLAIM is set, rejects any JWT whose "groups"
claim doesn't contain the configured group name. Requests that the JWT auth
middleware already whitelisted (e.g. /api/v1/evaluation) are passed through
unchecked, since they may have no JWT at all.
*/
func setupJWTRequireGroupClaimMiddleware() *requireGroupClaim {
	return &requireGroupClaim{
		Group: Config.JWTAuthRequireGroupClaim,
	}
}

type requireGroupClaim struct {
	Group string
}

func (c *requireGroupClaim) checkGroups(r *http.Request) bool {
	if whiteListed, _ := r.Context().Value(whiteListedCtxKey{}).(bool); whiteListed {
		return true
	}

	token, ok := r.Context().Value(Config.JWTAuthUserProperty).(*jwt.Token)
	if !ok {
		return false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		groups := util.SafeStringSlice(claims["groups"])
		for _, s := range groups {
			if s == c.Group {
				return true
			}
		}
	}

	return false
}

func (c *requireGroupClaim) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !c.checkGroups(r) {
		jwtErrorHandler(w, r, "Not member of authorized group")
		return
	}
	next(w, r)
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

	if slices.Contains(a.ExactWhitelistPaths, path) {
		return true
	}

	for _, p := range a.PrefixWhitelistPaths {
		if p != "" && util.HasSafePrefix(path, p) {
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
