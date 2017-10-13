package pprof

import . "net/http/pprof"
import "net/http"
import "strings"

// Create pprof middleware.
func New() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			pref := "/debug/pprof/"

			if strings.HasPrefix(path, pref) {
				name := strings.TrimPrefix(path, pref)
				handle(name)(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// Handle `typ`.
func handle(typ string) http.HandlerFunc {
	switch typ {
	case "profile":
		return Profile
	case "cmdline":
		return Cmdline
	case "symbol":
		return Symbol
	default:
		return Index
	}
}
