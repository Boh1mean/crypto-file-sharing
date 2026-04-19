package container

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"time"
)

type EncryptedFileContainer struct {
	Version               string    `json:"version"`
	SenderID              string    `json:"sender_id"`
	RecipientID           string    `json:"recipient_id"`
	EphemeralPublicKey    []byte    `json:"ephemeral_public_key"`
	KeyDerivationSalt     []byte    `json:"key_derivation_salt"`
	WrappedFileKeyNonce   []byte    `json:"wrapped_file_key_nonce"`
	WrappedFileKey        []byte    `json:"wrapped_file_key"`
	FileNonce             []byte    `json:"file_nonce"`
	Ciphertext            []byte    `json:"ciphertext"`
	FileHash              []byte    `json:"file_hash"`
	Metadata              Metadata  `json:"metadata"`
	Signature             []byte    `json:"signature"`
	SignatureAlgorithm    string    `json:"signature_algorithm"`
	KeyAgreementAlgorithm string    `json:"key_agreement_algorithm"`
	HashAlgorithm         string    `json:"hash_algorithm"`
	CreatedAt             time.Time `json:"created_at"`
}

type Metadata struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

type BuildInput struct {
	SenderID                  string
	RecipientID               string
	SenderSigningPrivateKey   *ecdsa.PrivateKey
	RecipientEncryptionPubKey *ecdh.PublicKey
	Plaintext                 []byte
	Metadata                  Metadata
}

type DecryptInput struct {
	Container                  EncryptedFileContainer
	RecipientEncryptionPrivKey *ecdh.PrivateKey
	SenderSigningPublicKey     *ecdsa.PublicKey
}

type DecryptOutput struct {
	Plaintext []byte
	Metadata  Metadata
}
