package crypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func TestGenerateECDH(t *testing.T) {
	priv, pub, err := GenerateECDH()
	if err != nil {
		t.Fatalf("generate ecdh key pair: %v", err)
	}

	if priv == nil {
		t.Fatal("private key must not be nil")
	}

	if pub == nil {
		t.Fatal("public key must not be nil")
	}
}

func TestGenerateECDSA(t *testing.T) {
	priv, err := GenerateECDSA()
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}

	if priv == nil {
		t.Fatal("private key must not be nil")
	}
}

func TestSharedSecret_SameOnBothSides(t *testing.T) {
	alicePriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate alice key: %v", err)
	}

	bobPriv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate bob key: %v", err)
	}

	aliceSecret, err := SharedSecret(bobPriv.PublicKey(), alicePriv)
	if err != nil {
		t.Fatalf("alice shared secret: %v", err)
	}

	bobSecret, err := SharedSecret(alicePriv.PublicKey(), bobPriv)
	if err != nil {
		t.Fatalf("bob shared secret: %v", err)
	}

	if len(aliceSecret) == 0 {
		t.Fatal("shared secret is empty")
	}

	if !bytes.Equal(aliceSecret, bobSecret) {
		t.Fatal("shared secrets must be equal")
	}
}

func TestHashSHA256(t *testing.T) {
	data := []byte("hello world")

	got := HashSHA256(data)
	wantArr := sha256.Sum256(data)
	want := wantArr[:]

	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected hash: got %x, want %x", got, want)
	}
}

func TestSign(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}

	sig, err := Sign(priv, []byte("message"))
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}

	if len(sig) == 0 {
		t.Fatal("signature must not be empty")
	}
}

func TestVerify_ValidSignature(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}

	msg := []byte("message")
	hash := sha256.Sum256(msg)

	sig, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}

	if !Verify(&priv.PublicKey, msg, sig) {
		t.Fatal("expected signature to be valid")
	}
}

func TestVerify_InvalidSignatureForModifiedMessage(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}

	msg := []byte("message")
	hash := sha256.Sum256(msg)

	sig, err := ecdsa.SignASN1(rand.Reader, priv, hash[:])
	if err != nil {
		t.Fatalf("sign message: %v", err)
	}

	if Verify(&priv.PublicKey, []byte("tampered"), sig) {
		t.Fatal("expected signature to be invalid for modified message")
	}
}

func TestGenerateHKDF_Deterministic(t *testing.T) {
	sharedSecret := []byte("shared-secret")
	salt := []byte("salt")
	info := []byte("info")

	key1, err := GenerateHKDF(sharedSecret, salt, info, 32)
	if err != nil {
		t.Fatalf("generate hkdf key1: %v", err)
	}

	key2, err := GenerateHKDF(sharedSecret, salt, info, 32)
	if err != nil {
		t.Fatalf("generate hkdf key2: %v", err)
	}

	if len(key1) != 32 {
		t.Fatalf("unexpected key length: got %d, want 32", len(key1))
	}

	if !bytes.Equal(key1, key2) {
		t.Fatal("hkdf output must be deterministic for same inputs")
	}
}

func TestGenerateHKDF_DifferentInfoProducesDifferentKeys(t *testing.T) {
	sharedSecret := []byte("shared-secret")
	salt := []byte("salt")

	key1, err := GenerateHKDF(sharedSecret, salt, []byte("info-1"), 32)
	if err != nil {
		t.Fatalf("generate hkdf key1: %v", err)
	}

	key2, err := GenerateHKDF(sharedSecret, salt, []byte("info-2"), 32)
	if err != nil {
		t.Fatalf("generate hkdf key2: %v", err)
	}

	if bytes.Equal(key1, key2) {
		t.Fatal("hkdf keys must differ when info differs")
	}
}

func TestEncryptDecryptAESGCM_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	plaintext := []byte("secret message")
	aad := []byte("authenticated data")

	nonce, ciphertext, err := EncryptAESGCM(key, plaintext, aad)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if len(nonce) == 0 {
		t.Fatal("nonce must not be empty")
	}

	if len(ciphertext) == 0 {
		t.Fatal("ciphertext must not be empty")
	}

	decrypted, err := DecryptAESGCM(key, nonce, ciphertext, aad)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("unexpected plaintext: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptAESGCM_FailsWithWrongAAD(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate key: %v", err)
	}

	nonce, ciphertext, err := EncryptAESGCM(key, []byte("secret"), []byte("aad"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	_, err = DecryptAESGCM(key, nonce, ciphertext, []byte("wrong-aad"))
	if err == nil {
		t.Fatal("expected decrypt to fail with wrong aad")
	}
}

func TestEncryptAESGCM_InvalidKeySize(t *testing.T) {
	_, _, err := EncryptAESGCM([]byte("short"), []byte("plaintext"), nil)
	if err == nil {
		t.Fatal("expected error for invalid AES key size")
	}
}

func TestDecryptAESGCM_InvalidKeySize(t *testing.T) {
	_, err := DecryptAESGCM([]byte("short"), []byte("nonce"), []byte("ciphertext"), nil)
	if err == nil {
		t.Fatal("expected error for invalid AES key size")
	}
}
