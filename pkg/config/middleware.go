package config

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gohttp/pprof"
	"github.com/meatballhat/negroni-logrus"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"github.com/yadvendar/negroni-newrelic-go-agent"
)

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	n := negroni.New()

	if Config.CORSEnabled {
		n.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type"},
			ExposedHeaders: []string{"Www-Authenticate"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		}))
	}

	if Config.StatsdEnabled {
		n.Use(&statsdMiddleware{StatsdClient: Global.StatsdClient})
	}

	if Config.NewRelicEnabled {
		n.Use(&negroninewrelic.Newrelic{Application: &Global.NewrelicApp})
	}

	if Config.JWTAuthEnabled {
		n.Use(setupJWTAuthMiddleware())
	}

	if Config.MiddlewareVerboseLoggerEnabled {
		n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
	}

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewStatic(http.Dir("./browser/flagr-ui/dist/")))

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}

	return n
}

/**
setupJWTAuthMiddleware setup an JWTMiddleware from the ENV config
*/
func setupJWTAuthMiddleware() *auth {
	var signingMethod jwt.SigningMethod
	var validationKey interface{}
	var errParsingKey error

	switch Config.JWTAuthSigningMethod {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
		validationKey = []byte(Config.JWTAuthSecret)
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
		validationKey, errParsingKey = jwt.ParseRSAPublicKeyFromPEM([]byte(Config.JWTAuthSecret))
	default:
		signingMethod = jwt.SigningMethodHS256
		validationKey = []byte("")
	}

	return &auth{
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

type auth struct {
	PrefixWhitelistPaths []string
	ExactWhitelistPaths  []string
	JWTMiddleware        *jwtmiddleware.JWTMiddleware
}

func (a *auth) whitelist(req *http.Request) bool {
	path := req.URL.Path

	// If we set to 401 unauthorized, let the client handles the 401 itself
	if Config.JWTAuthNoTokenStatusCode == http.StatusUnauthorized {
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

func (a *auth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if a.whitelist(req) {
		next(w, req)
		return
	}
	a.JWTMiddleware.HandlerWithNext(w, req, next)
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
