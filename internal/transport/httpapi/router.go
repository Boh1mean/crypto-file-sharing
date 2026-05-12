package httpapi

import (
	"net/http"
	"time"

	"cryptocore/internal/service"
)

const (
	smallBodyLimit int64 = 4 * 1024          // 4 KB — для auth и регистрации
	largeBodyLimit int64 = 100 * 1024 * 1024 // 100 MB — для загрузки файлов
)

func NewRouter(users *service.UserService, files *service.FileService, auth *service.AuthService) http.Handler {
	handler := NewHandler(users, files)
	authHandler := NewAuthHandler(auth)
	requireAuth := RequireAuth(auth)

	// Rate limiting: не более 10 запросов в минуту с одного IP на auth-эндпоинтах.
	authRateLimit := RateLimit(10, time.Minute)

	mux := http.NewServeMux()

	// Публичные маршруты
	mux.Handle("POST /users",
		LimitBody(smallBodyLimit)(http.HandlerFunc(handler.CreateUser)))
	mux.Handle("POST /auth/challenge",
		authRateLimit(LimitBody(smallBodyLimit)(http.HandlerFunc(authHandler.CreateChallenge))))
	mux.Handle("POST /auth/verify",
		authRateLimit(LimitBody(smallBodyLimit)(http.HandlerFunc(authHandler.VerifyChallenge))))

	// Защищённые маршруты
	mux.Handle("GET /users/{id}",
		requireAuth(http.HandlerFunc(handler.GetUserPublicKeys)))
	mux.Handle("GET /users/by-username/{username}",
		requireAuth(http.HandlerFunc(handler.GetUserByUsername)))
	mux.Handle("POST /files",
		requireAuth(LimitBody(largeBodyLimit)(http.HandlerFunc(handler.StoreContainer))))
	mux.Handle("GET /files/{id}",
		requireAuth(http.HandlerFunc(handler.LoadContainer)))

	// SecurityHeaders оборачивает весь mux — применяется ко всем маршрутам.
	return SecurityHeaders(mux)
}
