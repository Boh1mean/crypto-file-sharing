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
	Username            string
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
		Username:            input.Username,
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
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/users/%d", id), nil, &out); err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	return decodePublicKeysResponse(out.ID, out.EncryptionPublicKey, out.SigningPublicKey)
}

func (c *Client) GetUserPublicKeysByUsername(ctx context.Context, username string) (GetUserPublicKeysOutput, error) {
	var out getUserByUsernameResponse
	if err := c.doJSON(ctx, http.MethodGet, "/users/by-username/"+username, nil, &out); err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	return decodePublicKeysResponse(out.ID, out.EncryptionPublicKey, out.SigningPublicKey)
}

func decodePublicKeysResponse(id int, encKeyB64, sigKeyB64 string) (GetUserPublicKeysOutput, error) {
	encryptionPubRaw, err := base64.StdEncoding.DecodeString(encKeyB64)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	encryptionPub, err := crypto.ParseECDHPublicKey(encryptionPubRaw)
	if err != nil {
		return GetUserPublicKeysOutput{}, err
	}

	signingPubRaw, err := base64.StdEncoding.DecodeString(sigKeyB64)
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
		ID:                  id,
		EncryptionPublicKey: encryptionPub,
		SigningPublicKey:    signingPub,
	}, nil
}
