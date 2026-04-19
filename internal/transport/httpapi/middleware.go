package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"cryptocore/internal/service"
)

// RequireAuth middleware — проверяет session token из заголовка Authorization: Bearer <token>.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				writeError(w, http.StatusUnauthorized, errors.New("authorization token required"))
				return
			}

			if _, err := auth.ValidateSession(r.Context(), token); err != nil {
				writeError(w, http.StatusUnauthorized, errors.New("invalid or expired session"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
