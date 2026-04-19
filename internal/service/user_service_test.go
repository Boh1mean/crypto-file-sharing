package service

import (
	"context"
	"crypto/x509"
	"cryptocore/internal/core/crypto"
	"cryptocore/internal/domain"
	"cryptocore/internal/infrastructure/memory"
	"cryptocore/internal/repository"
	"errors"
	"testing"
)

func TestUserService_CreateAndGetPublicKeys_Success(t *testing.T) {
	ctx := context.Background()

	_, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate ecdh key pair: %v", err)
	}

	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate ecdsa key pair: %v", err)
	}

	users := memory.NewUserRepository()
	service := NewUserService(users)

	out, err := service.CreateUser(ctx, domain.CreateUserInput{
		Username:            "alice",
		EncryptionPublicKey: encryptionPub,
		SigningPublicKey:    &signingPriv.PublicKey,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if out.ID == 0 {
		t.Fatal("expected non-zero generated ID")
	}

	keys, err := service.GetUserPublicKeys(ctx, domain.GetUserPublicKeysInput{ID: out.ID})
	if err != nil {
		t.Fatalf("get user public keys: %v", err)
	}

	if string(keys.EncryptionPublicKey.Bytes()) != string(encryptionPub.Bytes()) {
		t.Fatal("unexpected encryption public key")
	}

	gotSigning, err := x509.MarshalPKIXPublicKey(keys.SigningPublicKey)
	if err != nil {
		t.Fatalf("marshal loaded signing public key: %v", err)
	}
	wantSigning, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		t.Fatalf("marshal original signing public key: %v", err)
	}
	if string(gotSigning) != string(wantSigning) {
		t.Fatal("unexpected signing public key")
	}
}

func TestUserService_GetUserPublicKeys_FailsWhenUserNotFound(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepository()
	service := NewUserService(users)

	_, err := service.GetUserPublicKeys(ctx, domain.GetUserPublicKeysInput{ID: 1})
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
