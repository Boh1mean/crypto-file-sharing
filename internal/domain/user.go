package domain

import (
	"crypto/ecdh"
	"crypto/ecdsa"
)

type User struct {
	ID                  int
	Username            string
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey     *ecdsa.PublicKey
}

type CreateUserInput struct {
	Username            string
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey     *ecdsa.PublicKey
}

type CreateUserOutput struct {
	ID int
}

type GetUserPublicKeysInput struct {
	ID int
}

type GetUserPublicKeysOutput struct {
	ID                  int
	Username            string
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey     *ecdsa.PublicKey
}

type GetUserByUsernameInput struct {
	Username string
}

type GetUserByUsernameOutput struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey     *ecdsa.PublicKey
}
