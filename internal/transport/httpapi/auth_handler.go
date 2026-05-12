package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"cryptocore/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// CreateChallenge godoc
// POST /auth/challenge
// Тело: { "user_id": 1 }
// Ответ: { "nonce": "<base64>" }
func (h *AuthHandler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	var req createChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	nonce, err := h.auth.CreateChallenge(r.Context(), req.UserID)
	if err != nil {
		writeMappedAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, createChallengeResponse{
		Nonce: base64.StdEncoding.EncodeToString(nonce),
	})
}

// VerifyChallenge godoc
// POST /auth/verify
// Тело: { "user_id": 1, "signature": "<base64>" }
// Ответ: { "session_token": "...", "expires_at": "..." }
func (h *AuthHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	var req verifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid signature encoding"))
		return
	}

	session, err := h.auth.VerifyChallenge(r.Context(), req.UserID, signature)
	if err != nil {
		writeMappedAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, verifyChallengeResponse{
		SessionToken: session.Token,
		ExpiresAt:    session.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func writeMappedAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrChallengeExpired):
		slog.Warn("auth: challenge expired or not found")
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrInvalidSignature):
		slog.Warn("auth: invalid challenge signature")
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrSessionExpired):
		slog.Warn("auth: session expired")
		writeError(w, http.StatusUnauthorized, err)
	case errors.Is(err, service.ErrSessionNotFound):
		slog.Warn("auth: session not found")
		writeError(w, http.StatusUnauthorized, err)
	default:
		writeMappedError(w, err)
	}
}
