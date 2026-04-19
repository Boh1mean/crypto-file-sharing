package app

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"strings"

	clientapi "cryptocore/internal/client/api"
	"cryptocore/internal/client/keystore"
	"cryptocore/internal/core/crypto"
)

type LoginService struct {
	store *keystore.Store
}

func NewLoginService(store *keystore.Store) *LoginService {
	return &LoginService{store: store}
}

type LoginOutput struct {
	SessionToken string
}

// Login выполняет challenge-response аутентификацию.
// Если сервер не знает пользователя (например, после рестарта in-memory хранилища),
// автоматически перерегистрирует существующие публичные ключи и повторяет вход.
func (s *LoginService) Login(ctx context.Context) (LoginOutput, error) {
	identity, err := s.store.LoadIdentity()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("load identity: %w", err)
	}

	client := clientapi.NewClient(identity.ServerURL)

	challengeOut, err := client.RequestChallenge(ctx, identity.UserID)
	if err != nil {
		if isUserNotFound(err) {
			// Сервер потерял пользователя (in-memory рестарт) — переотправляем публичные ключи.
			if reErr := s.reRegisterPublicKeys(ctx, client, identity); reErr != nil {
				return LoginOutput{}, fmt.Errorf("re-register after server reset: %w", reErr)
			}
			// Повторяем challenge после успешной перерегистрации.
			challengeOut, err = client.RequestChallenge(ctx, identity.UserID)
			if err != nil {
				return LoginOutput{}, fmt.Errorf("request challenge after re-register: %w", err)
			}
		} else {
			return LoginOutput{}, fmt.Errorf("request challenge: %w", err)
		}
	}

	signature, err := crypto.Sign(identity.SigningPrivateKey, challengeOut.Nonce)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("sign challenge: %w", err)
	}

	verifyOut, err := client.VerifyChallenge(ctx, identity.UserID, signature)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("verify challenge: %w", err)
	}

	profile, err := s.store.Load()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("load profile: %w", err)
	}

	profile.SessionToken = verifyOut.SessionToken
	profile.TokenExpiresAt = verifyOut.ExpiresAt

	if err := s.store.Save(profile); err != nil {
		return LoginOutput{}, fmt.Errorf("save profile: %w", err)
	}

	return LoginOutput{SessionToken: verifyOut.SessionToken}, nil
}

// reRegisterPublicKeys отправляет на сервер существующие публичные ключи из профиля.
// Вызывается когда сервер потерял пользователя (in-memory рестарт), но ключи сохранены локально.
func (s *LoginService) reRegisterPublicKeys(ctx context.Context, client *clientapi.Client, identity keystore.Identity) error {
	profile, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("load profile: %w", err)
	}

	encryptionPub, err := parseECDHPublicKey(profile.EncryptionPublicKey, identity)
	if err != nil {
		return fmt.Errorf("parse encryption public key: %w", err)
	}

	signingPub, err := parseECDSAPublicKey(profile.SigningPublicKeyPKIX)
	if err != nil {
		return fmt.Errorf("parse signing public key: %w", err)
	}

	_, err = client.CreateUser(ctx, clientapi.CreateUserInput{
		Username:            identity.Username,
		EncryptionPublicKey: encryptionPub,
		SigningPublicKey:    signingPub,
	})
	return err
}

// Logout удаляет session token из локального профиля.
func (s *LoginService) Logout(ctx context.Context) error {
	profile, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("load profile: %w", err)
	}

	profile.SessionToken = ""
	profile.TokenExpiresAt = profile.TokenExpiresAt.Local().AddDate(-1, 0, 0)

	return s.store.Save(profile)
}

func isUserNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "user not found")
}

func parseECDHPublicKey(raw []byte, identity keystore.Identity) (*ecdh.PublicKey, error) {
	// Восстанавливаем публичный ключ из приватного (он всегда доступен).
	return identity.EncryptionPrivateKey.PublicKey(), nil
}

func parseECDSAPublicKey(pkixRaw []byte) (*ecdsa.PublicKey, error) {
	parsed, err := x509.ParsePKIXPublicKey(pkixRaw)
	if err != nil {
		return nil, err
	}
	pub, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unexpected key type")
	}
	return pub, nil
}
