package config

import (
	"bytes"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gohttp/pprof"
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
		n.Use(negronilogrus.NewMiddlewareFromLogger(logrus.StandardLogger(), "flagr"))
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
			AllowedOrigins:   []string{"*"},
			AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "Authorization", "Time_Zone"},
			ExposedHeaders:   []string{"Www-Authenticate"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			AllowCredentials: true,
		}))
	}

	if Config.JWTAuthEnabled {
		n.Use(setupJWTAuthMiddleware())
	}

	if Config.BasicAuthEnabled {
		n.Use(setupBasicAuthMiddleware())
	}

	n.Use(&negroni.Static{
		Dir:       http.Dir("./browser/flagr-ui/dist/"),
		Prefix:    Config.WebPrefix,
		IndexFile: "index.html",
	})

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

type OidcConfiguation struct {
	JwkURI string `json:"jwks_uri"` // This is the only field we care about for now.
}

type OidcJwk struct {
	JWKeyID     string   `json:"kid"` // JWTs will have a "kid" that will match one of these
	SigningKeys []string `json:"x5c"` // We prefer the x5c since it's easier
	Exponent    string   `json:"e"`   // The least preferred way is recreating the key through exponent.
	Modulus     string   `json:"n"`   // The modulus
}

type OidcJwksConfiguration struct {
	Jwks []OidcJwk `json:"keys"`
}

func discoverOidcJwk(tokenKeyID string) *OidcJwk {
	// First we need to access the well known url to find out what the jwk_uri is.
	resp, err := http.Get(Config.JWTAuthOIDCWellKnownURL)
	if err != nil {
		logrus.Errorln("Failed to contact OIDC well known URL")
		return nil
	}

	// Read in the oidc config information, and unmarshal it into our oidc config.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("Failed to read from OIDC well known URL")
		return nil
	}

	var oidcConfig OidcConfiguation
	json.Unmarshal([]byte(body), &oidcConfig)

	if oidcConfig.JwkURI == "" {
		logrus.Errorln("OIDC configuration didn't contain a valid jwks_uri")
		return nil
	}

	// Now we need to read in the jwks info.
	oidcJwksResp, err := http.Get(oidcConfig.JwkURI)
	if err != nil {
		logrus.Errorln("Failed to contact the JWK URI")
		return nil
	}

	// Read in the jwks and unmarshall it.
	defer oidcJwksResp.Body.Close()
	body, err = ioutil.ReadAll(oidcJwksResp.Body)
	if err != nil {
		logrus.Errorln("Failed to read the JWKs body")
		return nil
	}

	var oidcJwks OidcJwksConfiguration

	json.Unmarshal([]byte(body), &oidcJwks)

	// Find the key that matches the one in the JWT
	if oidcJwks.Jwks == nil {
		logrus.Errorln("Missing keys in the jwk config")
		return nil
	}

	var numKeys = len(oidcJwks.Jwks)
	if numKeys == 0 {
		logrus.Errorln("No keys in the jwk config")
		return nil
	}

	matchingKey := -1
	for currentKey := 0; currentKey < numKeys; currentKey++ {
		if oidcJwks.Jwks[currentKey].JWKeyID == tokenKeyID {
			matchingKey = currentKey
			break
		}
	}

	if matchingKey == -1 {
		logrus.Errorln("No matching key found in the jwk config")
		return nil
	}

	return &oidcJwks.Jwks[matchingKey]
}

func extractJWTSigningKeyFromX5C(oidcJwk *OidcJwk) (interface{}, error) {
	correctKey := oidcJwk.SigningKeys[0]

	// Now we will take the first key and convert it into a cert
	// that the library we user are familiar with. This cert requires
	// a very specific format that is dependent on whitespace :/
	// There can only be 64 characters on a line or else it won't read it
	// so we have to do that manually. Also the BEGIN and END lines need to
	// be their own lines.
	numCharsInKey := len(correctKey)
	var numLines int = numCharsInKey / 64
	if numLines%64 > 0 {
		numLines++
	}

	var jwtCert string = "-----BEGIN PUBLIC KEY-----\n"
	for currentLine := 0; currentLine < numLines; currentLine++ {
		startCharacter := currentLine * 64
		endCharacter := startCharacter + 64
		if startCharacter+64 > numCharsInKey {
			endCharacter = numCharsInKey
		}

		jwtCert += correctKey[startCharacter:endCharacter] + "\n"
	}
	jwtCert += "-----END PUBLIC KEY-----"

	return jwt.ParseRSAPublicKeyFromPEM([]byte(jwtCert))
}

func calculateJWTSigningKey(jwk *OidcJwk) (interface{}, error) {
	if jwk.Modulus == "" || jwk.Exponent == "" {
		return "", errors.New("Invalid Modulus or Exponent provided")
	}

	// Decode the modulus and move it into a big int.
	logrus.Printf("Modulus found: " + jwk.Modulus)
	decN, err := base64.RawURLEncoding.DecodeString(jwk.Modulus)
	if err != nil {
		logrus.Errorf("Failed to decode modulus string.")
		return "", err
	}
	n := big.NewInt(0)
	n.SetBytes(decN)

	// Decode the exponent
	logrus.Printf("Exponent found: " + jwk.Exponent)
	eStr := jwk.Exponent
	decE, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		logrus.Errorf("Failed to decode exponent string")
		return "", err
	}

	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		logrus.Errorf("Failed to read exponent bytes")
		return "", err
	}
	pKey := &rsa.PublicKey{N: n, E: int(e)}
	logrus.Printf("Created public key.")
	return pKey, nil
}

/**
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

	var validationKeyGetter = func(token *jwt.Token) (interface{}, error) {
		return validationKey, errParsingKey
	}

	if Config.JWTAuthOIDCWellKnownURL != "" {
		validationKeyGetter = func(token *jwt.Token) (interface{}, error) {

			// If this is truly an OIDC token, it should have a "kid" header in the token.
			var tokenKeyID = token.Header["kid"]
			if tokenKeyID == nil {
				return "", errors.New("Missing key id in the JWT")
			}

			var jwk = discoverOidcJwk(tokenKeyID.(string))
			if jwk == nil {
				return "", errors.New("Failed to find JWK for JWT")
			}

			if len(jwk.SigningKeys) > 0 {
				return extractJWTSigningKeyFromX5C(jwk)
			} else if jwk.Exponent != "" && jwk.Modulus != "" {
				return calculateJWTSigningKey(jwk)
			} else {
				return "", errors.New("JWK is invalid")
			}
		}
	}

	return &jwtAuth{
		PrefixWhitelistPaths: Config.JWTAuthPrefixWhitelistPaths,
		ExactWhitelistPaths:  Config.JWTAuthExactWhitelistPaths,
		JWTMiddleware: jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: validationKeyGetter,
			SigningMethod:       signingMethod,
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

func (a *jwtAuth) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if a.whitelist(req) {
		next(w, req)
		return
	}
	a.JWTMiddleware.HandlerWithNext(w, req, next)
}

/**
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
