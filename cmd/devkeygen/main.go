package main

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"

	"cryptocore/internal/core/crypto"
)

type createUserRequest struct {
	ID                  int    `json:"id"`
	EncryptionPublicKey string `json:"encryption_public_key"`
	SigningPublicKey    string `json:"signing_public_key"`
}

func main() {
	userID := flag.Int("id", 1, "user id")
	flag.Parse()

	_, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		log.Fatal(err)
	}

	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		log.Fatal(err)
	}

	signingPubRaw, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	payload := createUserRequest{
		ID:                  *userID,
		EncryptionPublicKey: base64.StdEncoding.EncodeToString(encryptionPub.Bytes()),
		SigningPublicKey:    base64.StdEncoding.EncodeToString(signingPubRaw),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		log.Fatal(err)
	}
}
