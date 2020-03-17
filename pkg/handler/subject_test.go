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

	defer func() { config.Config.JWTAuthEnabled = false }()
	config.Config.JWTAuthEnabled = true

	r, _ := http.NewRequest("GET", "", nil)
	assert.Equal(t, getSubjectFromRequest(r), "")

	ctx = context.TODO()
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "")

	//nolint:staticcheck // jwt-middleware is using the string type of context key
	ctx = context.WithValue(ctx, config.Config.JWTAuthUserProperty, &jwt.Token{})
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "")

	//nolint:staticcheck // jwt-middleware is using the string type of context key
	ctx = context.WithValue(ctx, config.Config.JWTAuthUserProperty, &jwt.Token{
		Claims: jwt.MapClaims{
			"sub": "foo@example.com",
		},
		Valid: true,
	})
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "foo@example.com")
}

func TestGetSubjectFromOauthProxy(t *testing.T) {
	var ctx = context.Background()

	defer func() { config.Config.HeaderAuthEnabled = false }()
	config.Config.HeaderAuthEnabled = true

	r, _ := http.NewRequest("GET", "", nil)
	assert.Equal(t, getSubjectFromRequest(r), "")

	r.Header.Set(config.Config.HeaderAuthUserField, "foo@example.com")
	assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "foo@example.com")
}
