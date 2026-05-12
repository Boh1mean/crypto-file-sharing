package httpapi

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"cryptocore/internal/core/crypto"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
	"cryptocore/internal/service"
)

type Handler struct {
	users *service.UserService
	files *service.FileService
}

func NewHandler(users *service.UserService, files *service.FileService) *Handler {
	return &Handler{
		users: users,
		files: files,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	encryptionPublicKey, err := decodeECDHPublicKey(req.EncryptionPublicKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	signingPublicKey, err := decodeECDSAPublicKey(req.SigningPublicKey)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	out, err := h.users.CreateUser(r.Context(), domain.CreateUserInput{
		Username:            req.Username,
		EncryptionPublicKey: encryptionPublicKey,
		SigningPublicKey:    signingPublicKey,
	})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createUserResponse{ID: out.ID})
}

func (h *Handler) GetUserPublicKeys(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	out, err := h.users.GetUserPublicKeys(r.Context(), domain.GetUserPublicKeysInput{ID: id})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	encryptionPublicKey := base64.StdEncoding.EncodeToString(out.EncryptionPublicKey.Bytes())
	signingPublicKey, err := x509.MarshalPKIXPublicKey(out.SigningPublicKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, getUserPublicKeysResponse{
		ID:                  out.ID,
		Username:            out.Username,
		EncryptionPublicKey: encryptionPublicKey,
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPublicKey),
	})
}

func (h *Handler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		writeError(w, http.StatusBadRequest, errors.New("username is required"))
		return
	}

	out, err := h.users.GetUserByUsername(r.Context(), domain.GetUserByUsernameInput{Username: username})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	encryptionPublicKey := base64.StdEncoding.EncodeToString(out.EncryptionPublicKey.Bytes())
	signingPublicKey, err := x509.MarshalPKIXPublicKey(out.SigningPublicKey)
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, getUserByUsernameResponse{
		ID:                  out.ID,
		EncryptionPublicKey: encryptionPublicKey,
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPublicKey),
	})
}

func (h *Handler) StoreContainer(w http.ResponseWriter, r *http.Request) {
	senderID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	var req storeContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	containerBytes, err := base64.StdEncoding.DecodeString(req.Container)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	out, err := h.files.StoreContainer(r.Context(), domain.StoreContainerInput{
		SenderID:       senderID,
		RecipientID:    req.RecipientID,
		ContainerBytes: containerBytes,
		FileName:       req.FileName,
		MimeType:       req.MimeType,
		Size:           req.Size,
	})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, storeContainerResponse{ID: out.ID})
}

func (h *Handler) LoadContainer(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	id, err := parseIDParam(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	out, err := h.files.LoadContainer(r.Context(), domain.LoadContainerInput{ID: id})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	if out.RecipientID != userID {
		slog.Warn("access denied: user is not the file recipient",
			"user_id", userID,
			"file_id", id,
			"recipient_id", out.RecipientID,
		)
		writeError(w, http.StatusForbidden, errors.New("access denied"))
		return
	}

	writeJSON(w, http.StatusOK, loadContainerResponse{
		ID:          out.ID,
		SenderID:    out.SenderID,
		RecipientID: out.RecipientID,
		Container:   base64.StdEncoding.EncodeToString(out.ContainerBytes),
		FileName:    out.FileName,
		MimeType:    out.MimeType,
		Size:        out.Size,
		CreatedAt:   out.CreatedAt.UTC().Format(http.TimeFormat),
	})
}

func parseIDParam(raw string) (int, error) {
	return strconv.Atoi(raw)
}

func decodeECDHPublicKey(encoded string) (*ecdh.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return crypto.ParseECDHPublicKey(raw)
}

func decodeECDSAPublicKey(encoded string) (*ecdsa.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	parsed, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, err
	}

	publicKey, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("signing public key is not ECDSA")
	}

	return publicKey, nil
}

func writeMappedError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrFileNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrContainerNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrUserAlreadyExists):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, repository.ErrFileAlreadyExists):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusRequestTimeout, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	if status >= http.StatusInternalServerError {
		slog.Error("internal server error", "err", err)
		msg = "internal server error"
	}
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
