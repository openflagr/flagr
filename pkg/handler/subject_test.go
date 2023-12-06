package handler

import (
	"context"
	"net/http"
	"testing"

	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/openflagr/flagr/pkg/config"
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

func TestGetSubjectFromCookie(t *testing.T) {
	var ctx = context.Background()
	defer func() { config.Config.CookieAuthEnabled = false }()
	config.Config.CookieAuthEnabled = true

	t.Run("test HS256 happy codepath", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "", nil)
		assert.Equal(t, getSubjectFromRequest(r), "")

		r.AddCookie(&http.Cookie{
			Name:  config.Config.CookieAuthUserField,
			Value: "eyJhbGciOiJIUzI1NiIsImtpZCI6IjEyMzQ1In0.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZW1haWwiOiJhYmNAZXhhbXBsZS5jb20iLCJpYXQiOjE1MTYyMzkwMjJ9.tzRXenFic8Eqg2awzO0eiX6Rozy_mmsJVzLJfUUfREI",
		})
		assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "abc@example.com")
	})

	t.Run("test HS256 empty claim", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "", nil)
		assert.Equal(t, getSubjectFromRequest(r), "")

		r.AddCookie(&http.Cookie{
			Name:  config.Config.CookieAuthUserField,
			Value: "eyJhbGciOiJIUzI1NiIsImtpZCI6IjEyMzQ1In0.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyfQ.C_YsEkcHa7aSVQILzJAayFgJk-sj1cmNWIWUm7m7vy4",
		})
		assert.Equal(t, getSubjectFromRequest(r.WithContext(ctx)), "")
	})
}
