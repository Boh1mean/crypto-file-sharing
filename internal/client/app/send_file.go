package app

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	clientapi "cryptocore/internal/client/api"
	"cryptocore/internal/client/keystore"
	"cryptocore/internal/core/container"
)

type SendFileService struct {
	store *keystore.Store
}

func NewSendFileService(store *keystore.Store) *SendFileService {
	return &SendFileService{store: store}
}

type SendFileInput struct {
	RecipientUsername string
	FilePath          string
}

type SendFileOutput struct {
	FileID int
}

func (s *SendFileService) Send(ctx context.Context, input SendFileInput) (SendFileOutput, error) {
	identity, err := s.store.LoadIdentity()
	if err != nil {
		return SendFileOutput{}, err
	}

	plaintext, err := os.ReadFile(input.FilePath)
	if err != nil {
		return SendFileOutput{}, err
	}

	apiClient := clientapi.NewClient(identity.ServerURL).WithToken(identity.SessionToken)
	recipientKeys, err := apiClient.GetUserPublicKeysByUsername(ctx, input.RecipientUsername)
	if err != nil {
		return SendFileOutput{}, err
	}

	fileName := filepath.Base(input.FilePath)
	mimeType := http.DetectContentType(plaintext)

	encryptedContainer, err := container.BuildContainer(container.BuildInput{
		SenderID:                  identity.Username,
		RecipientID:               input.RecipientUsername,
		SenderSigningPrivateKey:   identity.SigningPrivateKey,
		RecipientEncryptionPubKey: recipientKeys.EncryptionPublicKey,
		Plaintext:                 plaintext,
		Metadata: container.Metadata{
			FileName: fileName,
			MimeType: mimeType,
			Size:     int64(len(plaintext)),
		},
	})
	if err != nil {
		return SendFileOutput{}, err
	}

	rawContainer, err := container.Marshal(*encryptedContainer)
	if err != nil {
		return SendFileOutput{}, err
	}

	storeOut, err := apiClient.StoreContainer(ctx, clientapi.StoreContainerInput{
		SenderID:       identity.UserID,
		RecipientID:    recipientKeys.ID,
		ContainerBytes: rawContainer,
		FileName:       fileName,
		MimeType:       mimeType,
		Size:           int64(len(plaintext)),
	})
	if err != nil {
		return SendFileOutput{}, err
	}

	return SendFileOutput{FileID: storeOut.ID}, nil
}
