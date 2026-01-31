package middleware

import (
	"net/http"
	"strings"
)

// CORS is a minimal CORS middleware suitable for browser-based SPAs.
//
// origins can be:
// - "*" to allow any origin (recommended only when not using cookies)
// - comma-separated list of allowed origins (e.g. "https://app.example.com,https://staging.example.com")
func CORS(origins string) func(http.Handler) http.Handler {
	allowedAll := strings.TrimSpace(origins) == "*" || strings.TrimSpace(origins) == ""

	allowed := map[string]struct{}{}
	if !allowedAll {
		for _, o := range strings.Split(origins, ",") {
			origin := strings.TrimSpace(o)
			if origin != "" {
				allowed[origin] = struct{}{}
			}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowedAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					if _, ok := allowed[origin]; ok {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Add("Vary", "Origin")
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
