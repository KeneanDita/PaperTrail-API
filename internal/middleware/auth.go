package middleware

import (
	"context"
	"net/http"
	"strings"

	"papertrail/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userContextKey contextKey = "user"

// UserContext holds minimal data extracted from the JWT for downstream handlers.
type UserContext struct {
	Subject  string
	Role     string
	RawToken string
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				utils.RespondError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			tokenString := strings.TrimPrefix(header, "Bearer ")

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				utils.RespondError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				utils.RespondError(w, http.StatusUnauthorized, "invalid claims")
				return
			}

			role, _ := claims["role"].(string)
			sub, _ := claims["sub"].(string)

			ctx := context.WithValue(r.Context(), userContextKey, &UserContext{
				Subject:  sub,
				Role:     role,
				RawToken: tokenString,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *UserContext {
	val, _ := ctx.Value(userContextKey).(*UserContext)
	return val
}
