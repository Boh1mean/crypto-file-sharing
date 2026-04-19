package app

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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
	FileID      int
	RecipientID int
	FilePath    string
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
	recipientKeys, err := apiClient.GetUserPublicKeys(ctx, input.RecipientID)
	if err != nil {
		return SendFileOutput{}, err
	}

	fileName := filepath.Base(input.FilePath)
	mimeType := http.DetectContentType(plaintext)

	encryptedContainer, err := container.BuildContainer(container.BuildInput{
		SenderID:                  strconv.Itoa(identity.UserID),
		RecipientID:               strconv.Itoa(input.RecipientID),
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

	_, err = apiClient.StoreContainer(ctx, clientapi.StoreContainerInput{
		ID:             input.FileID,
		SenderID:       identity.UserID,
		RecipientID:    input.RecipientID,
		ContainerBytes: rawContainer,
		FileName:       fileName,
		MimeType:       mimeType,
		Size:           int64(len(plaintext)),
	})
	if err != nil {
		return SendFileOutput{}, err
	}

	return SendFileOutput{FileID: input.FileID}, nil
}
