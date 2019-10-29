package handler

import (
	"net/http"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/pkg/util"

	jwt "github.com/dgrijalva/jwt-go"
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

		} else if config.Config.AuthProxyEnabled {
			return r.Header.Get(config.Config.AuthProxyHeader)
		}

	return ""
}
