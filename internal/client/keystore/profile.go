package keystore

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"time"
)

type Profile struct {
	ServerURL            string    `json:"server_url"`
	UserID               int       `json:"user_id"`
	EncryptionPrivateKey []byte    `json:"encryption_private_key"`
	SigningPrivateKey     []byte    `json:"signing_private_key"`
	EncryptionPublicKey  []byte    `json:"encryption_public_key"`
	SigningPublicKeyPKIX []byte    `json:"signing_public_key_pkix"`
	SessionToken         string    `json:"session_token,omitempty"`
	TokenExpiresAt       time.Time `json:"token_expires_at,omitempty"`
}

// HasValidSession возвращает true если сохранённый токен ещё не истёк.
func (p Profile) HasValidSession() bool {
	return p.SessionToken != "" && time.Now().Before(p.TokenExpiresAt)
}

type Identity struct {
	ServerURL            string
	UserID               int
	EncryptionPrivateKey *ecdh.PrivateKey
	SigningPrivateKey     *ecdsa.PrivateKey
	SessionToken         string
}
