package transport

import (
	"net/http"

	"cryptocore/internal/service"
	"cryptocore/internal/transport/httpapi"
)

func NewRouter(users *service.UserService, files *service.FileService, auth *service.AuthService) http.Handler {
	return httpapi.NewRouter(users, files, auth)
}
