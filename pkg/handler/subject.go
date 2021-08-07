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
	}

	return ""
}
