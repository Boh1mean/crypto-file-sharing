package app

import (
	"context"
	"crypto/x509"
	"fmt"

	clientapi "cryptocore/internal/client/api"
	"cryptocore/internal/client/keystore"
	"cryptocore/internal/core/crypto"
)

type RegisterUserService struct {
	store *keystore.Store
}

func NewRegisterUserService(store *keystore.Store) *RegisterUserService {
	return &RegisterUserService{store: store}
}

type RegisterUserInput struct {
	Username string
}

type RegisterUserOutput struct {
	UserID       int
	SessionToken string
}

func (s *RegisterUserService) Register(ctx context.Context, input RegisterUserInput) (RegisterUserOutput, error) {
	encryptionPriv, encryptionPub, err := crypto.GenerateECDH()
	if err != nil {
		return RegisterUserOutput{}, err
	}

	signingPriv, err := crypto.GenerateECDSA()
	if err != nil {
		return RegisterUserOutput{}, err
	}

	client := clientapi.NewClient(ServerURL)
	out, err := client.CreateUser(ctx, clientapi.CreateUserInput{
		Username:            input.Username,
		EncryptionPublicKey: encryptionPub,
		SigningPublicKey:    &signingPriv.PublicKey,
	})
	if err != nil {
		return RegisterUserOutput{}, err
	}

	signingPrivRaw, err := x509.MarshalECPrivateKey(signingPriv)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	signingPubRaw, err := x509.MarshalPKIXPublicKey(&signingPriv.PublicKey)
	if err != nil {
		return RegisterUserOutput{}, err
	}

	if err := s.store.Save(keystore.Profile{
		ServerURL:            ServerURL,
		UserID:               out.ID,
		Username:             input.Username,
		EncryptionPrivateKey: encryptionPriv.Bytes(),
		SigningPrivateKey:    signingPrivRaw,
		EncryptionPublicKey:  encryptionPub.Bytes(),
		SigningPublicKeyPKIX: signingPubRaw,
	}); err != nil {
		return RegisterUserOutput{}, err
	}

	loginSvc := NewLoginService(s.store)
	loginOut, err := loginSvc.Login(ctx)
	if err != nil {
		return RegisterUserOutput{}, fmt.Errorf("auto-login after registration: %w", err)
	}

	return RegisterUserOutput{
		UserID:       out.ID,
		SessionToken: loginOut.SessionToken,
	}, nil
}
