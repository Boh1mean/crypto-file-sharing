package api

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"

	"cryptocore/internal/core/crypto"
)

type CreateUserInput struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey    *ecdsa.PublicKey
}

type CreateUserOutput struct {
	ID int
}

type GetUserPublicKeysOutput struct {
	ID                  int
	EncryptionPublicKey *ecdh.PublicKey
	SigningPublicKey    *ecdsa.PublicKey
}

func (c *Client) CreateUser(ctx context.Context, input CreateUserInput) (CreateUserOutput, error) {
	signingPubRaw, err := x509.MarshalPKIXPublicKey(input.SigningPublicKey)
	if err != nil {
		return CreateUserOutput{}, err
	}

	var out createUserResponse
	err = c.doJSON(ctx, http.MethodPost, "/users", createUserRequest{
		ID:                  input.ID,
		EncryptionPublicKey: base64.StdEncoding.EncodeToString(input.EncryptionPublicKey.Bytes()),
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPubRaw),
	}, &out)
	if err != nil {
		return CreateUserOutput{}, err
	}

	return CreateUserOutput{ID: out.ID}, nil
}

func (c *Client) GetUserPublicKeys(ctx context.Context, id int) (GetUserPublicKeysOutput, error) {
	var out getUserPublicKeysResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/users/%d/public-keys", id), nil, &out); err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	encryptionPubRaw, err := base64.StdEncoding.DecodeString(out.EncryptionPublicKey)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	encryptionPub, err := crypto.ParseECDHPublicKey(encryptionPubRaw)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	signingPubRaw, err := base64.StdEncoding.DecodeString(out.SigningPublicKey)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	parsedSigningPub, err := x509.ParsePKIXPublicKey(signingPubRaw)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	signingPub, ok := parsedSigningPub.(*ecdsa.PublicKey)
	if !ok {
		return GetUserPublicKeysOutput{}, fmt.Errorf("unexpected signing public key type")
	}

	return GetUserPublicKeysOutput{
		ID:                  out.ID,
		EncryptionPublicKey: encryptionPub,
		SigningPublicKey:    signingPub,
	}, nil
}
