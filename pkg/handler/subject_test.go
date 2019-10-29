
package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/checkr/flagr/pkg/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestGetSubjectFromJWT(t *testing.T) {
	var ctx context.Context

	config.Config.JWTAuthEnabled = true

	r, _ := http.NewRequest("GET", "", nil)
	assert.Equal(t, getSubjectFromRequest(r), "")

	ctx = context.TODO()
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "")

	ctx = context.WithValue(ctx, interface{}(config.Config.JWTAuthUserProperty), &jwt.Token{})
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "")

	ctx = context.WithValue(ctx, interface{}(config.Config.JWTAuthUserProperty), &jwt.Token{
		Claims: jwt.MapClaims{
			"sub": "foo@example.com",
		},
		Valid: true,
	})
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "foo@example.com")
}

func TestGetSubjectFromOauthProxy(t *testing.T) {
	var ctx = context.Background()

	config.Config.JWTAuthEnabled = false
	config.Config.AuthProxyEnabled = true

	r, _ := http.NewRequest("GET", "", nil)
	assert.Equal(t, getSubjectFromRequest(r), "")

	r.Header.Set("X-Email", "foo@example.com")
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "foo@example.com")
}
