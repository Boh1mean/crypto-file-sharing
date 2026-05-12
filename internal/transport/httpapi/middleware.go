package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"cryptocore/internal/service"
)

type contextKey int

const userIDKey contextKey = iota

// UserIDFromContext извлекает userID из контекста запроса, установленного RequireAuth.
func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok
}

// RequireAuth проверяет Bearer-токен и кладёт userID в контекст запроса.
func RequireAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				slog.Warn("auth: missing token",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				writeError(w, http.StatusUnauthorized, errors.New("authorization token required"))
				return
			}

			userID, err := auth.ValidateSession(r.Context(), token)
			if err != nil {
				slog.Warn("auth: invalid or expired token",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"err", err,
				)
				writeError(w, http.StatusUnauthorized, errors.New("invalid or expired session"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SecurityHeaders добавляет базовые security-заголовки к каждому ответу.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

// LimitBody ограничивает максимальный размер тела запроса.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimiter — простой фиксированный оконный счётчик запросов по IP.
type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*windowState
	limit   int
	window  time.Duration
}

type windowState struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string]*windowState),
		limit:   limit,
		window:  window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	state, ok := rl.clients[ip]
	if !ok || now.After(state.resetAt) {
		rl.clients[ip] = &windowState{count: 1, resetAt: now.Add(rl.window)}
		return true
	}

	if state.count >= rl.limit {
		return false
	}
	state.count++
	return true
}

// RateLimit возвращает middleware, ограничивающий число запросов с одного IP.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !rl.allow(ip) {
				slog.Warn("rate limit exceeded",
					"ip", ip,
					"path", r.URL.Path,
				)
				writeError(w, http.StatusTooManyRequests, errors.New("too many requests, please slow down"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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
