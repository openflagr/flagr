package config

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gohttp/pprof"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	negroninewrelic "github.com/yadvendar/negroni-newrelic-go-agent"
)

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	pwd, _ := os.Getwd()
	n := negroni.New()

	if Config.CORSEnabled {
		n.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Accepts"},
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
		n.Use(&auth{
			WhitelistPaths: strings.Split(Config.JWTAuthWhitelistPaths, ","),
			JWTMiddleware: jwtmiddleware.New(jwtmiddleware.Options{
				ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
					return []byte(Config.JWTAuthSecret), nil
				},
				SigningMethod: jwt.SigningMethodHS256,
				Extractor: func(r *http.Request) (string, error) {
					c, err := r.Cookie(Config.JWTAuthCookieTokenName)
					if err != nil {
						return "", err
					}
					return c.Value, nil
				},
				UserProperty: Config.JWTAuthUserProperty,
				Debug:        Config.JWTAuthDebug,
				ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
					http.Redirect(w, r, Config.JWTAuthNoTokenRedirectURL, 307)
				},
			}),
		})
	}

	if Config.MiddlewareVerboseLoggerEnabled {
		n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
	}

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewStatic(http.Dir(pwd + "/browser/flagr-ui/dist/")))

	if Config.PProfEnabled {
		n.UseHandler(pprof.New()(handler))
	} else {
		n.UseHandler(handler)
	}

	return n
}

type auth struct {
	WhitelistPaths []string
	JWTMiddleware  *jwtmiddleware.JWTMiddleware
}

func (a *auth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	path := req.URL.Path
	for _, p := range a.WhitelistPaths {
		if p != "" && strings.HasPrefix(path, p) {
			next(w, req)
			return
		}
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
