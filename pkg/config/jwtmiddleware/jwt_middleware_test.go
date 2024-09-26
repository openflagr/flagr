package jwtmiddleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/negroni"
)

// defaultAuthorizationHeaderName is the default header name where the Auth
// token should be written
const defaultAuthorizationHeaderName = "Authorization"

// userPropertyName is the property name that will be set in the request context
const userPropertyName = "custom-user-property"

// the bytes read from the keys/sample-key file
// private key generated with http://kjur.github.io/jsjws/tool_jwt.html
var privateKey []byte

// TestUnauthenticatedRequest will perform requests with no Authorization header
func TestUnauthenticatedRequest(t *testing.T) {
	Convey("Simple unauthenticated request", t, func() {
		Convey("Unauthenticated GET to / path should return a 200 response", func() {
			w := makeUnauthenticatedRequest("GET", "/")
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Unauthenticated GET to /protected path should return a 401 response", func() {
			w := makeUnauthenticatedRequest("GET", "/protected")
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

// TestAuthenticatedRequest will perform requests with an Authorization header
func TestAuthenticatedRequest(t *testing.T) {
	var e error
	privateKey, e = readPrivateKey()
	if e != nil {
		panic(e)
	}
	Convey("Simple authenticated requests", t, func() {
		Convey("Authenticated GET to / path should return a 200 response", func() {
			w := makeAuthenticatedRequest("GET", "/", jwt.MapClaims{"foo": "bar"}, nil)
			So(w.Code, ShouldEqual, http.StatusOK)
		})
		Convey("Authenticated GET to /protected path should return a 200 response if expected algorithm is not specified", func() {
			var expectedAlgorithm jwt.SigningMethod = nil
			w := makeAuthenticatedRequest("GET", "/protected", jwt.MapClaims{"foo": "bar"}, expectedAlgorithm)
			So(w.Code, ShouldEqual, http.StatusOK)
			responseBytes, err := ioutil.ReadAll(w.Body)
			if err != nil {
				panic(err)
			}
			responseString := string(responseBytes)
			// check that the encoded data in the jwt was properly returned as json
			So(responseString, ShouldEqual, `{"text":"bar"}`)
		})
		Convey("Authenticated GET to /protected path should return a 200 response if expected algorithm is correct", func() {
			expectedAlgorithm := jwt.SigningMethodHS256
			w := makeAuthenticatedRequest("GET", "/protected", jwt.MapClaims{"foo": "bar"}, expectedAlgorithm)
			So(w.Code, ShouldEqual, http.StatusOK)
			responseBytes, err := ioutil.ReadAll(w.Body)
			if err != nil {
				panic(err)
			}
			responseString := string(responseBytes)
			// check that the encoded data in the jwt was properly returned as json
			So(responseString, ShouldEqual, `{"text":"bar"}`)
		})
		Convey("Authenticated GET to /protected path should return a 401 response if algorithm is not expected one", func() {
			expectedAlgorithm := jwt.SigningMethodRS256
			w := makeAuthenticatedRequest("GET", "/protected", jwt.MapClaims{"foo": "bar"}, expectedAlgorithm)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
			responseBytes, err := ioutil.ReadAll(w.Body)
			if err != nil {
				panic(err)
			}
			responseString := string(responseBytes)
			// check that the encoded data in the jwt was properly returned as json
			So(strings.TrimSpace(responseString), ShouldEqual, "Expected RS256 signing method but token specified HS256")
		})
	})
}

func makeUnauthenticatedRequest(method string, url string) *httptest.ResponseRecorder {
	return makeAuthenticatedRequest(method, url, nil, nil)
}

func makeAuthenticatedRequest(method string, url string, c jwt.Claims, expectedSignatureAlgorithm jwt.SigningMethod) *httptest.ResponseRecorder {
	r, _ := http.NewRequest(method, url, nil)
	if c != nil {
		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims = c
		// private key generated with http://kjur.github.io/jsjws/tool_jwt.html
		s, e := token.SignedString(privateKey)
		if e != nil {
			panic(e)
		}
		r.Header.Set(defaultAuthorizationHeaderName, fmt.Sprintf("bearer %v", s))
	}
	w := httptest.NewRecorder()
	n := createNegroniMiddleware(expectedSignatureAlgorithm)
	n.ServeHTTP(w, r)
	return w
}

func createNegroniMiddleware(expectedSignatureAlgorithm jwt.SigningMethod) *negroni.Negroni {
	// create a gorilla mux router for public requests
	publicRouter := mux.NewRouter().StrictSlash(true)
	publicRouter.Methods("GET").
		Path("/").
		Name("Index").
		Handler(http.HandlerFunc(indexHandler))

	// create a gorilla mux route for protected requests
	// the routes will be tested for jwt tokens in the default auth header
	protectedRouter := mux.NewRouter().StrictSlash(true)
	protectedRouter.Methods("GET").
		Path("/protected").
		Name("Protected").
		Handler(http.HandlerFunc(protectedHandler))
	// create a negroni handler for public routes
	negPublic := negroni.New()
	negPublic.UseHandler(publicRouter)

	// negroni handler for api request
	negProtected := negroni.New()
	//add the JWT negroni handler
	negProtected.Use(negroni.HandlerFunc(JWT(expectedSignatureAlgorithm).HandlerWithNext))
	negProtected.UseHandler(protectedRouter)

	//Create the main router
	mainRouter := mux.NewRouter().StrictSlash(true)

	mainRouter.Handle("/", negPublic)
	mainRouter.Handle("/protected", negProtected)
	//if routes match the handle prefix then I need to add this dummy matcher {_dummy:.*}
	mainRouter.Handle("/protected/{_dummy:.*}", negProtected)

	n := negroni.Classic()
	// This are the "GLOBAL" middlewares that will be applied to every request
	// examples are listed below:
	//n.Use(gzip.Gzip(gzip.DefaultCompression))
	//n.Use(negroni.HandlerFunc(SecurityMiddleware().HandlerFuncWithNext))
	n.UseHandler(mainRouter)

	return n
}

// JWT creates the middleware that parses a JWT encoded token
func JWT(expectedSignatureAlgorithm jwt.SigningMethod) *JWTMiddleware {
	return New(Options{
		Debug:               false,
		CredentialsOptional: false,
		UserProperty:        userPropertyName,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			if privateKey == nil {
				var err error
				privateKey, err = readPrivateKey()
				if err != nil {
					panic(err)
				}
			}
			return privateKey, nil
		},
		SigningMethod: expectedSignatureAlgorithm,
	})
}

// readPrivateKey will load the keys/sample-key file into the
// global privateKey variable
func readPrivateKey() ([]byte, error) {
	pvtKeyStr1 := "pvt-key-example"
	privateKey := []byte(pvtKeyStr1)
	return privateKey, nil
}

// indexHandler will return an empty 200 OK response
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// protectedHandler will return the content of the "foo" encoded data
// in the token as json -> {"text":"bar"}
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve the token from the context
	u := r.Context().Value(userPropertyName)
	user := u.(*jwt.Token)
	respondJSON(user.Claims.(jwt.MapClaims)["foo"].(string), w)
}

// Response quick n' dirty Response struct to be encoded as json
type Response struct {
	Text string `json:"text"`
}

// respondJSON will take an string to write through the writer as json
func respondJSON(text string, w http.ResponseWriter) {
	response := Response{text}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResponse)
}
