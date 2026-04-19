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

	senderID, _ := users.Create(ctx, domain.User{})
	recipientID, _ := users.Create(ctx, domain.User{})

	containerBytes := []byte(`{"version":"v1"}`)

	out, err := service.StoreContainer(ctx, domain.StoreContainerInput{
		SenderID:       senderID,
		RecipientID:    recipientID,
		ContainerBytes: containerBytes,
		FileName:       "hello.txt",
		MimeType:       "text/plain",
		Size:           12,
	})
	if err != nil {
		t.Fatalf("store container: %v", err)
	}
	if out.ID == 0 {
		t.Fatal("expected non-zero generated file ID")
	}

	record, err := files.GetByID(ctx, out.ID)
	if err != nil {
		t.Fatalf("load file record: %v", err)
	}
	if record.StorageKey == "" {
		t.Fatal("expected non-empty storage key")
	}

	storedContainer, err := containers.Get(ctx, record.StorageKey)
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

	recipientID, _ := users.Create(ctx, domain.User{})

	_, err := service.StoreContainer(ctx, domain.StoreContainerInput{
		SenderID:       99999,
		RecipientID:    recipientID,
		ContainerBytes: []byte(`{"version":"v1"}`),
		FileName:       "hello.txt",
		MimeType:       "text/plain",
		Size:           12,
	})
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFileService_LoadContainer_Success(t *testing.T) {
	ctx := context.Background()

	files := memory.NewFileRepository()
	containers := memory.NewContainerStorage()
	service := NewFileService(memory.NewUserRepository(), files, containers)

	containerBytes := []byte(`{"version":"v1"}`)
	storageKey := "files/test.container"

	fileID, err := files.Create(ctx, domain.FileRecord{
		SenderID:    1,
		RecipientID: 2,
		StorageKey:  storageKey,
		FileName:    "hello.txt",
		MimeType:    "text/plain",
		Size:        12,
	})
	if err != nil {
		t.Fatalf("seed file record: %v", err)
	}

	if err := containers.Save(ctx, storageKey, containerBytes); err != nil {
		t.Fatalf("seed container: %v", err)
	}

	out, err := service.LoadContainer(ctx, domain.LoadContainerInput{ID: fileID})
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

	fileID, err := files.Create(ctx, domain.FileRecord{
		SenderID:   1,
		StorageKey: "files/missing.container",
	})
	if err != nil {
		t.Fatalf("seed file record: %v", err)
	}

	_, err = service.LoadContainer(ctx, domain.LoadContainerInput{ID: fileID})
	if !errors.Is(err, repository.ErrContainerNotFound) {
		t.Fatalf("expected ErrContainerNotFound, got %v", err)
	}
}
