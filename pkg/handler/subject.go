package handler

import (
	"net/http"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/util"

	jwt "github.com/form3tech-oss/jwt-go"
)

func getSubjectFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	if config.Config.JWTAuthEnabled {
		token, ok := r.Context().Value(config.Config.JWTAuthUserProperty).(*jwt.Token)
		if !ok {
			return ""
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return util.SafeString(claims[config.Config.JWTAuthUserClaim])
		}

	} else if config.Config.HeaderAuthEnabled {
		return r.Header.Get(config.Config.HeaderAuthUserField)
	} else if config.Config.CookieAuthEnabled {
		c, err := r.Cookie(config.Config.CookieAuthUserField)
		if err != nil {
			return ""
		}
		if config.Config.CookieAuthUserFieldJWTClaim != "" {
			// for this case, we choose to skip the error check because just like HeaderAuthUserField
			// in the future, we can extend this function to support cookie jwt token validation
			// this assumes that the cookie we get already passed the auth middleware
			token, _ := jwt.Parse(c.Value, func(token *jwt.Token) (any, error) { return "", nil })
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				return util.SafeString(claims[config.Config.CookieAuthUserFieldJWTClaim])
			}
		}
		return c.Value
	}

	return ""
}
