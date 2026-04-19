package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type FileService struct {
	users      repository.UserRepository
	files      repository.FileRepository
	containers repository.ContainerStorage
}

func NewFileService(
	users repository.UserRepository,
	files repository.FileRepository,
	containers repository.ContainerStorage,
) *FileService {
	return &FileService{
		users:      users,
		files:      files,
		containers: containers,
	}
}

func (s *FileService) StoreContainer(
	ctx context.Context,
	input domain.StoreContainerInput,
) (domain.StoreContainerOutput, error) {
	if _, err := s.users.GetByID(ctx, input.SenderID); err != nil {
		return domain.StoreContainerOutput{}, err
	}

	if _, err := s.users.GetByID(ctx, input.RecipientID); err != nil {
		return domain.StoreContainerOutput{}, err
	}

	storageKey, err := generateStorageKey()
	if err != nil {
		return domain.StoreContainerOutput{}, fmt.Errorf("generate storage key: %w", err)
	}

	if err := s.containers.Save(ctx, storageKey, input.ContainerBytes); err != nil {
		return domain.StoreContainerOutput{}, err
	}

	record := domain.FileRecord{
		Size:        input.Size,
		SenderID:    input.SenderID,
		RecipientID: input.RecipientID,
		StorageKey:  storageKey,
		FileName:    input.FileName,
		MimeType:    input.MimeType,
		CreatedAt:   time.Now(),
	}

	id, err := s.files.Create(ctx, record)
	if err != nil {
		return domain.StoreContainerOutput{}, err
	}

	return domain.StoreContainerOutput{ID: id}, nil
}

func (s *FileService) LoadContainer(
	ctx context.Context,
	input domain.LoadContainerInput,
) (domain.LoadContainerOutput, error) {
	record, err := s.files.GetByID(ctx, input.ID)
	if err != nil {
		return domain.LoadContainerOutput{}, err
	}

	rawContainer, err := s.containers.Get(ctx, record.StorageKey)
	if err != nil {
		return domain.LoadContainerOutput{}, err
	}

	return domain.LoadContainerOutput{
		ID:             record.ID,
		SenderID:       record.SenderID,
		RecipientID:    record.RecipientID,
		ContainerBytes: rawContainer,
		FileName:       record.FileName,
		MimeType:       record.MimeType,
		Size:           record.Size,
		CreatedAt:      record.CreatedAt,
	}, nil
}

func generateStorageKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "files/" + hex.EncodeToString(b) + ".container", nil
}
