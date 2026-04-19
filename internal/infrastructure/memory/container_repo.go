package memory

import (
	"context"
	"cryptocore/internal/repository"
	"sync"
)

type ContainerStorage struct {
	mu         sync.RWMutex
	containers map[string][]byte
}

func NewContainerStorage() *ContainerStorage {
	return &ContainerStorage{
		containers: make(map[string][]byte),
	}
}

func (s *ContainerStorage) Save(ctx context.Context, key string, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)

	s.containers[key] = cp
	return nil
}

func (s *ContainerStorage) Get(ctx context.Context, key string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.containers[key]
	if !exists {
		return nil, repository.ErrContainerNotFound
	}

	cp := make([]byte, len(data))
	copy(cp, data)

	return cp, nil
}

func (s *ContainerStorage) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.containers[key]; !exists {
		return repository.ErrContainerNotFound
	}

	delete(s.containers, key)
	return nil
}

func (s *ContainerStorage) Exists(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.containers[key]
	return exists, nil
}
