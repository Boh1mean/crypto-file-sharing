package service

import (
	"context"
	"cryptocore/internal/domain"
	"cryptocore/internal/infrastructure/memory"
	"cryptocore/internal/repository"
	"errors"
	"testing"
)

func TestFileService_StoreContainer_Success(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepository()
	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(users, files, containers)

	if err := users.Create(ctx, domain.User{ID: 1}); err != nil {
		t.Fatalf("create sender: %v", err)
	}
	if err := users.Create(ctx, domain.User{ID: 2}); err != nil {
		t.Fatalf("create recipient: %v", err)
	}

	containerBytes := []byte(`{"version":"v1"}`)

	out, err := service.StoreContainer(ctx, domain.StoreContainerInput{
		ID:             10,
		SenderID:       1,
		RecipientID:    2,
		ContainerBytes: containerBytes,
		FileName:       "hello.txt",
		MimeType:       "text/plain",
		Size:           12,
	})
	if err != nil {
		t.Fatalf("store container: %v", err)
	}
	if out.ID != 10 {
		t.Fatalf("unexpected output id: got %d want 10", out.ID)
	}

	record, err := files.GetByID(ctx, 10)
	if err != nil {
		t.Fatalf("load file record: %v", err)
	}
	if record.StorageKey != makeStorageKey(10) {
		t.Fatalf("unexpected storage key: got %q", record.StorageKey)
	}

	storedContainer, err := containers.Get(ctx, makeStorageKey(10))
	if err != nil {
		t.Fatalf("load container: %v", err)
	}
	if string(storedContainer) != string(containerBytes) {
		t.Fatalf("unexpected container bytes: got %q want %q", storedContainer, containerBytes)
	}
}

func TestFileService_StoreContainer_FailsWhenSenderNotFound(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepository()
	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(users, files, containers)

	if err := users.Create(ctx, domain.User{ID: 2}); err != nil {
		t.Fatalf("create recipient: %v", err)
	}

	_, err := service.StoreContainer(ctx, domain.StoreContainerInput{
		ID:             10,
		SenderID:       1,
		RecipientID:    2,
		ContainerBytes: []byte(`{"version":"v1"}`),
		FileName:       "hello.txt",
		MimeType:       "text/plain",
		Size:           12,
	})
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFileService_StoreContainer_FailsWhenFileAlreadyExists(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepository()
	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(users, files, containers)

	if err := users.Create(ctx, domain.User{ID: 1}); err != nil {
		t.Fatalf("create sender: %v", err)
	}
	if err := users.Create(ctx, domain.User{ID: 2}); err != nil {
		t.Fatalf("create recipient: %v", err)
	}
	if err := files.Create(ctx, domain.FileRecord{ID: 10}); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	_, err := service.StoreContainer(ctx, domain.StoreContainerInput{
		ID:             10,
		SenderID:       1,
		RecipientID:    2,
		ContainerBytes: []byte(`{"version":"v1"}`),
		FileName:       "hello.txt",
		MimeType:       "text/plain",
		Size:           12,
	})
	if !errors.Is(err, repository.ErrFileAlreadyExists) {
		t.Fatalf("expected ErrFileAlreadyExists, got %v", err)
	}
}

func TestFileService_LoadContainer_Success(t *testing.T) {
	ctx := context.Background()

	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(memory.NewUserRepository(), files, containers)

	containerBytes := []byte(`{"version":"v1"}`)
	storageKey := makeStorageKey(10)

	if err := files.Create(ctx, domain.FileRecord{
		ID:          10,
		SenderID:    1,
		RecipientID: 2,
		StorageKey:  storageKey,
		FileName:    "hello.txt",
		MimeType:    "text/plain",
		Size:        12,
	}); err != nil {
		t.Fatalf("seed file record: %v", err)
	}
	if err := containers.Save(ctx, storageKey, containerBytes); err != nil {
		t.Fatalf("seed container: %v", err)
	}

	out, err := service.LoadContainer(ctx, domain.LoadContainerInput{ID: 10})
	if err != nil {
		t.Fatalf("load container: %v", err)
	}
	if string(out.ContainerBytes) != string(containerBytes) {
		t.Fatalf("unexpected container bytes: got %q want %q", out.ContainerBytes, containerBytes)
	}
	if out.FileName != "hello.txt" {
		t.Fatalf("unexpected file name: got %q", out.FileName)
	}
}

func TestFileService_LoadContainer_FailsWhenContainerMissing(t *testing.T) {
	ctx := context.Background()

	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(memory.NewUserRepository(), files, containers)

	if err := files.Create(ctx, domain.FileRecord{
		ID:         10,
		SenderID:   1,
		StorageKey: makeStorageKey(10),
	}); err != nil {
		t.Fatalf("seed file record: %v", err)
	}

	_, err := service.LoadContainer(ctx, domain.LoadContainerInput{ID: 10})
	if !errors.Is(err, repository.ErrContainerNotFound) {
		t.Fatalf("expected ErrContainerNotFound, got %v", err)
	}
}
