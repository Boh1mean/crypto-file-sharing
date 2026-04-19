package crypto

import (
	"crypto/ecdh"
	"crypto/ecdsa"
)

type Core interface {
	GenerateECDH() (*ecdh.PrivateKey, *ecdh.PublicKey, error)
	GenerateECDSA() (*ecdsa.PrivateKey, error)
	SharedSecret(pub *ecdh.PublicKey, priv *ecdh.PrivateKey) ([]byte, error)
	HashSHA256(data []byte) []byte
	RandomBytes(size int) ([]byte, error)
	Sign(priv *ecdsa.PrivateKey, msg []byte) ([]byte, error)
	Verify(pub *ecdsa.PublicKey, msg, sig []byte) bool
	GenerateHKDF(sharedSecret, salt, info []byte, size int) ([]byte, error)
	EncryptAESGCM(key, plaintext, aad []byte) (nonce, ciphertext []byte, err error)
	DecryptAESGCM(key, nonce, ciphertext, aad []byte) ([]byte, error)
	ParseECDHPublicKey(data []byte) (*ecdh.PublicKey, error)
}
