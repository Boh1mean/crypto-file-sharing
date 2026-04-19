package service

import (
	"context"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
	"fmt"
	"time"
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

	exists, err := s.files.Exists(ctx, input.ID)
	if err != nil {
		return domain.StoreContainerOutput{}, err
	}
	if exists {
		return domain.StoreContainerOutput{}, repository.ErrFileAlreadyExists
	}

	storageKey := makeStorageKey(input.ID)
	if err := s.containers.Save(ctx, storageKey, input.ContainerBytes); err != nil {
		return domain.StoreContainerOutput{}, err
	}

	record := domain.FileRecord{
		ID:          input.ID,
		Size:        input.Size,
		SenderID:    input.SenderID,
		RecipientID: input.RecipientID,
		StorageKey:  storageKey,
		FileName:    input.FileName,
		MimeType:    input.MimeType,
		CreatedAt:   time.Now(),
	}
	if err := s.files.Create(ctx, record); err != nil {
		return domain.StoreContainerOutput{}, err
	}

	return domain.StoreContainerOutput{ID: input.ID}, nil
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

func makeStorageKey(fileID int) string {
	return fmt.Sprintf("files/%d.container", fileID)
}
