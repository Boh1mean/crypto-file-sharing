package domain

import (
	"crypto/ecdh"
	"crypto/ecdsa"
)

type User struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey    *ecdsa.PublicKey
}

type CreateUserInput struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey    *ecdsa.PublicKey
}

type CreateUserOutput struct {
	ID int
}

type GetUserPublicKeysInput struct {
	ID int
}

type GetUserPublicKeysOutput struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey    *ecdsa.PublicKey
}
