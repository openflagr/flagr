package config

import (
	"net/http"
	"os"
	"strings"

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

	n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
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
		if strings.HasPrefix(path, p) {
			next(w, req)
			return
		}
	}
	a.JWTMiddleware.HandlerWithNext(w, req, next)
}
