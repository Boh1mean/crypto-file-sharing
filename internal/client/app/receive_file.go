package app

import (
	"context"
	"os"
	"path/filepath"

	clientapi "cryptocore/internal/client/api"
	"cryptocore/internal/client/keystore"
	"cryptocore/internal/core/container"
)

type ReceiveFileService struct {
	store *keystore.Store
}

func NewReceiveFileService(store *keystore.Store) *ReceiveFileService {
	return &ReceiveFileService{store: store}
}

type ReceiveFileInput struct {
	FileID    int
	OutputDir string
}

type ReceiveFileOutput struct {
	FileID         int
	OutputFilePath string
}

func (s *ReceiveFileService) Receive(ctx context.Context, input ReceiveFileInput) (ReceiveFileOutput, error) {
	identity, err := s.store.LoadIdentity()
	if err != nil {
		return ReceiveFileOutput{}, err
	}

	apiClient := clientapi.NewClient(identity.ServerURL).WithToken(identity.SessionToken)
	loadedContainer, err := apiClient.LoadContainer(ctx, input.FileID)
	if err != nil {
		return ReceiveFileOutput{}, err
	}

	senderKeys, err := apiClient.GetUserPublicKeys(ctx, loadedContainer.SenderID)
	if err != nil {
		return ReceiveFileOutput{}, err
	}

	encryptedContainer, err := container.Unmarshal(loadedContainer.ContainerBytes)
	if err != nil {
		return ReceiveFileOutput{}, err
	}

	decrypted, err := container.VerifyAndDecryptContainer(container.DecryptInput{
		Container:                  encryptedContainer,
		RecipientEncryptionPrivKey: identity.EncryptionPrivateKey,
		SenderSigningPublicKey:     senderKeys.SigningPublicKey,
	})
	if err != nil {
		return ReceiveFileOutput{}, err
	}

	if err := os.MkdirAll(input.OutputDir, 0o755); err != nil {
		return ReceiveFileOutput{}, err
	}

	outputFilePath := filepath.Join(input.OutputDir, decrypted.Metadata.FileName)
	if err := os.WriteFile(outputFilePath, decrypted.Plaintext, 0o600); err != nil {
		return ReceiveFileOutput{}, err
	}

	return ReceiveFileOutput{
		FileID:         input.FileID,
		OutputFilePath: outputFilePath,
	}, nil
}
