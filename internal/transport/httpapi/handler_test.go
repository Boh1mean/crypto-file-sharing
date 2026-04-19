package httpapi

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cryptocore/internal/core/crypto"
	"cryptocore/internal/infrastructure/memory"
	"cryptocore/internal/service"
)

func TestHTTPAPI_CreateUserAndGetPublicKeys_Success(t *testing.T) {
	router := newTestRouter()

	_, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate ecdh keys: %v", err)
	}

	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate ecdsa keys: %v", err)
	}

	signingPubRaw, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		t.Fatalf("marshal signing public key: %v", err)
	}

	createReq := createUserRequest{
		ID:                  1,
		EncryptionPublicKey: base64.StdEncoding.EncodeToString(encryptionPub.Bytes()),
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPubRaw),
	}

	rec := performJSONRequest(t, router, http.MethodPost, "/users", createReq, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var createResp createUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.ID != 1 {
		t.Fatalf("unexpected user id: got %d want 1", createResp.ID)
	}

	token := getSessionToken(t, router, 1, signingPriv)

	rec = performRequest(t, router, http.MethodGet, "/users/1/public-keys", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var keysResp getUserPublicKeysResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &keysResp); err != nil {
		t.Fatalf("decode keys response: %v", err)
	}

	if keysResp.EncryptionPublicKey != createReq.EncryptionPublicKey {
		t.Fatalf("unexpected encryption key: got %q want %q", keysResp.EncryptionPublicKey, createReq.EncryptionPublicKey)
	}
	if keysResp.SigningPublicKey != createReq.SigningPublicKey {
		t.Fatalf("unexpected signing key: got %q want %q", keysResp.SigningPublicKey, createReq.SigningPublicKey)
	}
}

func TestHTTPAPI_StoreAndLoadContainer_Success(t *testing.T) {
	router := newTestRouter()

	signingPriv1 := registerUser(t, router, 1)
	registerUser(t, router, 2)
	token := getSessionToken(t, router, 1, signingPriv1)

	container := []byte(`{"version":"v1","ciphertext":"abc"}`)
	storeReq := storeContainerRequest{
		ID:          10,
		SenderID:    1,
		RecipientID: 2,
		Container:   base64.StdEncoding.EncodeToString(container),
		FileName:    "hello.txt",
		MimeType:    "text/plain",
		Size:        31,
	}

	rec := performJSONRequest(t, router, http.MethodPost, "/files", storeReq, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	rec = performRequest(t, router, http.MethodGet, "/files/10", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var loadResp loadContainerResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &loadResp); err != nil {
		t.Fatalf("decode load response: %v", err)
	}

	if loadResp.ID != 10 {
		t.Fatalf("unexpected file id: got %d want 10", loadResp.ID)
	}
	if loadResp.Container != storeReq.Container {
		t.Fatalf("unexpected container: got %q want %q", loadResp.Container, storeReq.Container)
	}
	if loadResp.FileName != storeReq.FileName {
		t.Fatalf("unexpected file name: got %q want %q", loadResp.FileName, storeReq.FileName)
	}
}

func TestHTTPAPI_StoreContainer_FailsWhenRecipientMissing(t *testing.T) {
	router := newTestRouter()

	signingPriv1 := registerUser(t, router, 1)
	token := getSessionToken(t, router, 1, signingPriv1)

	rec := performJSONRequest(t, router, http.MethodPost, "/files", storeContainerRequest{
		ID:          10,
		SenderID:    1,
		RecipientID: 2,
		Container:   base64.StdEncoding.EncodeToString([]byte(`{"version":"v1"}`)),
		FileName:    "hello.txt",
		MimeType:    "text/plain",
		Size:        12,
	}, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestHTTPAPI_CreateUser_FailsWhenAlreadyExists(t *testing.T) {
	router := newTestRouter()

	registerUser(t, router, 1)

	_, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate ecdh keys: %v", err)
	}
	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate ecdsa keys: %v", err)
	}
	signingPubRaw, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		t.Fatalf("marshal signing public key: %v", err)
	}

	rec := performJSONRequest(t, router, http.MethodPost, "/users", createUserRequest{
		ID:                  1,
		EncryptionPublicKey: base64.StdEncoding.EncodeToString(encryptionPub.Bytes()),
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPubRaw),
	}, "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestHTTPAPI_CreateUser_FailsWhenKeyPayloadInvalid(t *testing.T) {
	router := newTestRouter()

	rec := performJSONRequest(t, router, http.MethodPost, "/users", createUserRequest{
		ID:                  1,
		EncryptionPublicKey: "not-base64",
		SigningPublicKey:    "not-base64",
	}, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got %d want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHTTPAPI_ProtectedRoute_FailsWithoutToken(t *testing.T) {
	router := newTestRouter()
	registerUser(t, router, 1)

	rec := performRequest(t, router, http.MethodGet, "/users/1/public-keys", nil, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

// --- helpers ---

func newTestRouter() http.Handler {
	userRepo := memory.NewUserRepository()
	fileRepo := memory.NewFileRepository()
	containerStorage := memory.NewContainerStorage()
	sessionRepo := memory.NewSessionRepository()
	challengeRepo := memory.NewChallengeRepository()

	userService := service.NewUserService(userRepo)
	fileService := service.NewFileService(userRepo, fileRepo, containerStorage)
	authService := service.NewAuthService(userRepo, sessionRepo, challengeRepo)

	return NewRouter(userService, fileService, authService)
}

// registerUser регистрирует пользователя и возвращает его signing private key для последующего логина.
func registerUser(t *testing.T, router http.Handler, id int) *ecdsa.PrivateKey {
	t.Helper()

	_, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate ecdh keys: %v", err)
	}
	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate ecdsa keys: %v", err)
	}
	signingPubRaw, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		t.Fatalf("marshal signing public key: %v", err)
	}

	rec := performJSONRequest(t, router, http.MethodPost, "/users", createUserRequest{
		ID:                  id,
		EncryptionPublicKey: base64.StdEncoding.EncodeToString(encryptionPub.Bytes()),
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPubRaw),
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register user failed: status=%d body=%s", rec.Code, rec.Body.String())
	}

	return signingPriv
}

// getSessionToken выполняет challenge-response и возвращает session token.
func getSessionToken(t *testing.T, router http.Handler, userID int, signingPriv *ecdsa.PrivateKey) string {
	t.Helper()

	// 1. Запрашиваем challenge.
	rec := performJSONRequest(t, router, http.MethodPost, "/auth/challenge", createChallengeRequest{
		UserID: userID,
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("create challenge failed: status=%d body=%s", rec.Code, rec.Body.String())
	}

	var challengeResp createChallengeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &challengeResp); err != nil {
		t.Fatalf("decode challenge response: %v", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(challengeResp.Nonce)
	if err != nil {
		t.Fatalf("decode nonce: %v", err)
	}

	// 2. Подписываем nonce.
	signature, err := crypto.Sign(signingPriv, nonce)
	if err != nil {
		t.Fatalf("sign nonce: %v", err)
	}

	// 3. Отправляем подпись.
	rec = performJSONRequest(t, router, http.MethodPost, "/auth/verify", verifyChallengeRequest{
		UserID:    userID,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("verify challenge failed: status=%d body=%s", rec.Code, rec.Body.String())
	}

	var verifyResp verifyChallengeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &verifyResp); err != nil {
		t.Fatalf("decode verify response: %v", err)
	}

	return verifyResp.SessionToken
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, payload any, token string) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func performRequest(t *testing.T, handler http.Handler, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
