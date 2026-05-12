package app

import (
	"context"
	"os"
	"path/filepath"
	"time"

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

type InboxItem struct {
	ID             int
	SenderUsername string
	FileName       string
	Size           int64
	CreatedAt      time.Time
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

func (s *ReceiveFileService) ListInbox(ctx context.Context) ([]InboxItem, error) {
	identity, err := s.store.LoadIdentity()
	if err != nil {
		return nil, err
	}

	apiClient := clientapi.NewClient(identity.ServerURL).WithToken(identity.SessionToken)
	apiItems, err := apiClient.ListInbox(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]InboxItem, len(apiItems))
	for i, a := range apiItems {
		items[i] = InboxItem{
			ID:             a.ID,
			SenderUsername: a.SenderUsername,
			FileName:       a.FileName,
			Size:           a.Size,
			CreatedAt:      a.CreatedAt,
		}
	}
	return items, nil
}
