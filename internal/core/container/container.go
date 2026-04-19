package container

import (
	"bytes"
	"crypto/ecdsa"
	"cryptocore/internal/core/crypto"
	"encoding/json"
	"errors"
	"time"
)

func Marshal(c EncryptedFileContainer) ([]byte, error) {
	return json.Marshal(c)
}

func Unmarshal(data []byte) (EncryptedFileContainer, error) {
	var c EncryptedFileContainer
	err := json.Unmarshal(data, &c)
	if err != nil {
		return EncryptedFileContainer{}, err
	}

	return c, nil
}

func canonicalBytesForSignature(c EncryptedFileContainer) ([]byte, error) {
	c.Signature = nil
	return json.Marshal(c)
}

func BuildContainer(input BuildInput) (*EncryptedFileContainer, error) {
	if len(input.Plaintext) == 0 {
		return nil, errors.New("plaintext is empty")
	}

	if input.SenderSigningPrivateKey == nil {
		return nil, errors.New("sender signing private key is nil")
	}

	if input.RecipientEncryptionPubKey == nil {
		return nil, errors.New("recipient encryption public key is nil")
	}

	fileKey, err := crypto.RandomBytes(32)
	if err != nil {
		return nil, err
	}
	fileHash := crypto.HashSHA256(input.Plaintext)

	aadStruct := map[string]interface{}{
		"sender_id":    input.SenderID,
		"recipient_id": input.RecipientID,
		"file_name":    input.Metadata.FileName,
		"mime_type":    input.Metadata.MimeType,
		"size":         input.Metadata.Size,
		"version":      "v1",
	}

	aad, err := json.Marshal(aadStruct)
	if err != nil {
		return nil, err
	}

	fileNonce, ciphertext, err := crypto.EncryptAESGCM(fileKey, input.Plaintext, aad)
	if err != nil {
		return nil, err
	}

	ephemeralPrivKey, ephemeralPubKey, err := crypto.GenerateECDH()
	if err != nil {
		return nil, err
	}

	sharedSecret, err := crypto.SharedSecret(input.RecipientEncryptionPubKey, ephemeralPrivKey)
	if err != nil {
		return nil, err
	}

	// 7. Генерируем salt для HKDF
	salt, err := crypto.RandomBytes(32)
	if err != nil {
		return nil, err
	}

	// 8. Выводим wrap key через HKDF
	wrapKey, err := crypto.GenerateHKDF(sharedSecret, salt, []byte("file-wrap-key"), 32)
	if err != nil {
		return nil, err
	}

	// 9. Шифруем сам fileKey через wrapKey
	wrappedFileKeyNonce, wrappedFileKey, err := crypto.EncryptAESGCM(wrapKey, fileKey, nil)
	if err != nil {
		return nil, err
	}

	container := &EncryptedFileContainer{
		Version:               "v1",
		SenderID:              input.SenderID,
		RecipientID:           input.RecipientID,
		EphemeralPublicKey:    ephemeralPubKey.Bytes(),
		KeyDerivationSalt:     salt,
		WrappedFileKeyNonce:   wrappedFileKeyNonce,
		WrappedFileKey:        wrappedFileKey,
		FileNonce:             fileNonce,
		Ciphertext:            ciphertext,
		FileHash:              fileHash,
		Metadata:              input.Metadata,
		Signature:             nil,
		SignatureAlgorithm:    "ECDSA-P256-SHA256",
		KeyAgreementAlgorithm: "ECDH-P256",
		HashAlgorithm:         "SHA-256",
		CreatedAt:             time.Now(),
	}

	// 11. Канонически сериализуем контейнер без подписи
	canonical, err := canonicalBytesForSignature(*container)
	if err != nil {
		return nil, err
	}

	// 12. Подписываем контейнер
	signature, err := crypto.Sign(input.SenderSigningPrivateKey, canonical)
	if err != nil {
		return nil, err
	}

	// 13. Кладем подпись в контейнер
	container.Signature = signature

	return container, nil
}

func VerifyContainerSignature(container EncryptedFileContainer, senderPub *ecdsa.PublicKey) error {
	if senderPub == nil {
		return errors.New("sender public key is nil")
	}

	canonical, err := canonicalBytesForSignature(container)
	if err != nil {
		return err
	}

	ok := crypto.Verify(senderPub, canonical, container.Signature)
	if !ok {
		return errors.New("invalid container signature")
	}

	return nil
}

func VerifyAndDecryptContainer(input DecryptInput) (*DecryptOutput, error) {
	if input.RecipientEncryptionPrivKey == nil {
		return nil, errors.New("recipient private key is nil")
	}

	if input.SenderSigningPublicKey == nil {
		return nil, errors.New("sender signing public key is nil")
	}

	// 1. Проверяем подпись
	if err := VerifyContainerSignature(input.Container, input.SenderSigningPublicKey); err != nil {
		return nil, err
	}

	// 2. Восстанавливаем ephemeral public key
	ephemeralPubKey, err := crypto.ParseECDHPublicKey(input.Container.EphemeralPublicKey)
	if err != nil {
		return nil, err
	}

	// 3. Восстанавливаем shared secret
	sharedSecret, err := crypto.SharedSecret(ephemeralPubKey, input.RecipientEncryptionPrivKey)
	if err != nil {
		return nil, err
	}

	// 4. Получаем wrap key через HKDF
	wrapKey, err := crypto.GenerateHKDF(
		sharedSecret,
		input.Container.KeyDerivationSalt,
		[]byte("file-wrap-key"),
		32,
	)
	if err != nil {
		return nil, err
	}

	// 5. Расшифровываем fileKey
	fileKey, err := crypto.DecryptAESGCM(
		wrapKey,
		input.Container.WrappedFileKeyNonce,
		input.Container.WrappedFileKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 6. Собираем тот же AAD, что и при шифровании
	aadStruct := map[string]interface{}{
		"sender_id":    input.Container.SenderID,
		"recipient_id": input.Container.RecipientID,
		"file_name":    input.Container.Metadata.FileName,
		"mime_type":    input.Container.Metadata.MimeType,
		"size":         input.Container.Metadata.Size,
		"version":      input.Container.Version,
	}

	aad, err := json.Marshal(aadStruct)
	if err != nil {
		return nil, err
	}

	// 7. Расшифровываем файл
	plaintext, err := crypto.DecryptAESGCM(
		fileKey,
		input.Container.FileNonce,
		input.Container.Ciphertext,
		aad,
	)
	if err != nil {
		return nil, err
	}

	// 8. Проверяем хэш
	hash := crypto.HashSHA256(plaintext)
	if !bytes.Equal(hash, input.Container.FileHash) {
		return nil, errors.New("file hash mismatch")
	}

	return &DecryptOutput{
		Plaintext: plaintext,
		Metadata:  input.Container.Metadata,
	}, nil
}
