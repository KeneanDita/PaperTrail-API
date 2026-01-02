package middleware

import (
	"net/http"

	"papertrail/internal/utils"
)

func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				utils.RespondError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}

			if len(allowed) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowed[user.Role]; !ok {
				utils.RespondError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
