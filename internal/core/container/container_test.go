package container

import (
	"encoding/json"
	"testing"

	"cryptocore/internal/core/crypto"
)

func TestBuildAndDecryptContainer_Success(t *testing.T) {
	// sender signing key
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	// recipient encryption key pair
	recipientPriv, recipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate recipient ecdh key pair: %v", err)
	}

	plaintext := []byte("hello secure world")
	metadata := Metadata{
		FileName: "hello.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: recipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	out, err := VerifyAndDecryptContainer(DecryptInput{
		Container:                  *container,
		RecipientEncryptionPrivKey: recipientPriv,
		SenderSigningPublicKey:     &senderSignPriv.PublicKey,
	})
	if err != nil {
		t.Fatalf("verify and decrypt: %v", err)
	}

	if string(out.Plaintext) != string(plaintext) {
		t.Fatalf("plaintext mismatch: got %q want %q", string(out.Plaintext), string(plaintext))
	}

	if out.Metadata.FileName != metadata.FileName {
		t.Fatalf("metadata filename mismatch: got %q want %q", out.Metadata.FileName, metadata.FileName)
	}
}

func TestVerifyContainerSignature_FailsWhenCiphertextTampered(t *testing.T) {
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	_, recipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate recipient ecdh key pair: %v", err)
	}

	plaintext := []byte("hello secure world")
	metadata := Metadata{
		FileName: "file.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: recipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	container.Ciphertext[0] ^= 0xFF

	err = VerifyContainerSignature(*container, &senderSignPriv.PublicKey)
	if err == nil {
		t.Fatal("expected signature verification to fail after ciphertext tampering")
	}
}

func TestVerifyAndDecryptContainer_FailsWithWrongRecipientKey(t *testing.T) {
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	_, correctRecipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate correct recipient pub: %v", err)
	}

	wrongRecipientPriv, _, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate wrong recipient key pair: %v", err)
	}

	plaintext := []byte("hello secure world")
	metadata := Metadata{
		FileName: "secret.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: correctRecipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	_, err = VerifyAndDecryptContainer(DecryptInput{
		Container:                  *container,
		RecipientEncryptionPrivKey: wrongRecipientPriv,
		SenderSigningPublicKey:     &senderSignPriv.PublicKey,
	})
	if err == nil {
		t.Fatal("expected decrypt to fail with wrong recipient private key")
	}
}

func TestVerifyAndDecryptContainer_FailsWhenCiphertextTampered(t *testing.T) {
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	recipientPriv, recipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate recipient ecdh key pair: %v", err)
	}

	plaintext := []byte("important file content")
	metadata := Metadata{
		FileName: "important.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: recipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	container.Ciphertext[0] ^= 0xAA

	_, err = VerifyAndDecryptContainer(DecryptInput{
		Container:                  *container,
		RecipientEncryptionPrivKey: recipientPriv,
		SenderSigningPublicKey:     &senderSignPriv.PublicKey,
	})
	if err == nil {
		t.Fatal("expected decrypt to fail after ciphertext tampering")
	}
}

func TestVerifyAndDecryptContainer_FailsWhenWrappedFileKeyTampered(t *testing.T) {
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	recipientPriv, recipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate recipient ecdh key pair: %v", err)
	}

	plaintext := []byte("wrapped key test")
	metadata := Metadata{
		FileName: "wrapped.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: recipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	container.WrappedFileKey[0] ^= 0x55

	_, err = VerifyAndDecryptContainer(DecryptInput{
		Container:                  *container,
		RecipientEncryptionPrivKey: recipientPriv,
		SenderSigningPublicKey:     &senderSignPriv.PublicKey,
	})
	if err == nil {
		t.Fatal("expected decrypt to fail after wrapped file key tampering")
	}
}

func TestContainer_JSONRoundTrip(t *testing.T) {
	senderSignPriv, err := crypto.GenerateECDSA()
	if err != nil {
		t.Fatalf("generate sender signing key: %v", err)
	}

	recipientPriv, recipientPub, err := crypto.GenerateECDH()
	if err != nil {
		t.Fatalf("generate recipient ecdh key pair: %v", err)
	}

	plaintext := []byte("json round trip")
	metadata := Metadata{
		FileName: "roundtrip.txt",
		MimeType: "text/plain",
		Size:     int64(len(plaintext)),
	}

	container, err := BuildContainer(BuildInput{
		SenderID:                  "alice",
		RecipientID:               "bob",
		SenderSigningPrivateKey:   senderSignPriv,
		RecipientEncryptionPubKey: recipientPub,
		Plaintext:                 plaintext,
		Metadata:                  metadata,
	})
	if err != nil {
		t.Fatalf("build container: %v", err)
	}

	raw, err := json.Marshal(container)
	if err != nil {
		t.Fatalf("marshal container: %v", err)
	}

	var decoded EncryptedFileContainer
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal container: %v", err)
	}

	out, err := VerifyAndDecryptContainer(DecryptInput{
		Container:                  decoded,
		RecipientEncryptionPrivKey: recipientPriv,
		SenderSigningPublicKey:     &senderSignPriv.PublicKey,
	})
	if err != nil {
		t.Fatalf("verify and decrypt after json round trip: %v", err)
	}

	if string(out.Plaintext) != string(plaintext) {
		t.Fatalf("plaintext mismatch after json round trip: got %q want %q", string(out.Plaintext), string(plaintext))
	}
}
