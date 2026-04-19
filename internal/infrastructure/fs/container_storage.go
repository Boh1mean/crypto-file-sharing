package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"cryptocore/internal/repository"
)

// ContainerStorage stores encrypted container blobs on the local filesystem.
// Each container is stored as a single file under baseDir using the storage key as a relative path.
type ContainerStorage struct {
	baseDir string
}

func NewContainerStorage(baseDir string) (*ContainerStorage, error) {
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, err
	}
	return &ContainerStorage{baseDir: baseDir}, nil
}

func (s *ContainerStorage) Save(ctx context.Context, key string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	path := s.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func (s *ContainerStorage) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.path(key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, repository.ErrContainerNotFound
		}
		return nil, err
	}

	return data, nil
}

func (s *ContainerStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	err := os.Remove(s.path(key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return repository.ErrContainerNotFound
		}
		return err
	}

	return nil
}

func (s *ContainerStorage) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	_, err := os.Stat(s.path(key))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (s *ContainerStorage) path(key string) string {
	return filepath.Join(s.baseDir, filepath.FromSlash(key))
}
