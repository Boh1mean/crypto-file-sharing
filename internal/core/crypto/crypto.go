package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

func GenerateECDH() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	privKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	return privKey, privKey.PublicKey(), nil
}

func GenerateECDSA() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func SharedSecret(pub *ecdh.PublicKey, priv *ecdh.PrivateKey) ([]byte, error) {
	keyExchange, err := priv.ECDH(pub)
	if err != nil {
		return nil, err
	}

	return keyExchange, nil
}

func HashSHA256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func RandomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Sign(priv *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	hash := sha256.Sum256(msg)
	signature, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		return nil, err
	}

	return signature, nil

}

func Verify(pub *ecdsa.PublicKey, msg, sig []byte) bool {
	hash := sha256.Sum256(msg)
	validation := ecdsa.VerifyASN1(pub, hash[:], sig)
	return validation
}

func GenerateHKDF(sharedSecret, salt, info []byte, size int) ([]byte, error) {
	key := make([]byte, size)

	kdf := hkdf.New(sha256.New, sharedSecret, salt, info)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncryptAESGCM(key, plaintext, aad []byte) (nonce, ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, aad)
	return nonce, ciphertext, nil
}

func DecryptAESGCM(key, nonce, ciphertext, aad []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, aad)
}

func ParseECDHPublicKey(data []byte) (*ecdh.PublicKey, error) {
	return ecdh.P256().NewPublicKey(data)
}
