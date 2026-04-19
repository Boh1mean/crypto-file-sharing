package httpapi

import (
	"net/http"

	"cryptocore/internal/service"
)

func NewRouter(users *service.UserService, files *service.FileService, auth *service.AuthService) http.Handler {
	handler := NewHandler(users, files)
	authHandler := NewAuthHandler(auth)
	requireAuth := RequireAuth(auth)

	mux := http.NewServeMux()

	// Публичные маршруты
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("POST /auth/challenge", authHandler.CreateChallenge)
	mux.HandleFunc("POST /auth/verify", authHandler.VerifyChallenge)

	// Защищённые маршруты
	mux.Handle("GET /users/{id}", requireAuth(http.HandlerFunc(handler.GetUserPublicKeys)))
	mux.Handle("GET /users/by-username/{username}", requireAuth(http.HandlerFunc(handler.GetUserByUsername)))
	mux.Handle("POST /files", requireAuth(http.HandlerFunc(handler.StoreContainer)))
	mux.Handle("GET /files/{id}", requireAuth(http.HandlerFunc(handler.LoadContainer)))

	return mux
}
