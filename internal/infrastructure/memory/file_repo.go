package memory

import (
	"context"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
	"sort"
	"sync"
	"sync/atomic"
)

type FileRepository struct {
	mu      sync.RWMutex
	files   map[int]domain.FileRecord
	counter atomic.Int64
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files: make(map[int]domain.FileRecord),
	}
}

func (r *FileRepository) Create(ctx context.Context, file domain.FileRecord) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := int(r.counter.Add(1))
	file.ID = id
	r.files[id] = file
	return id, nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int) (domain.FileRecord, error) {
	select {
	case <-ctx.Done():
		return domain.FileRecord{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	file, exists := r.files[id]
	if !exists {
		return domain.FileRecord{}, repository.ErrFileNotFound
	}

	return file, nil
}

func (r *FileRepository) Update(ctx context.Context, file domain.FileRecord) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.files[file.ID]; !exists {
		return repository.ErrFileNotFound
	}

	r.files[file.ID] = file
	return nil
}

func (r *FileRepository) Delete(ctx context.Context, id int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.files[id]; !exists {
		return repository.ErrFileNotFound
	}

	delete(r.files, id)
	return nil
}

func (r *FileRepository) Exists(ctx context.Context, id int) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.files[id]
	return exists, nil
}

func (r *FileRepository) List(ctx context.Context) ([]domain.FileRecord, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.FileRecord, 0, len(r.files))
	for _, file := range r.files {
		result = append(result, file)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (r *FileRepository) ListByRecipientID(ctx context.Context, recipientID int) ([]domain.FileRecord, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.FileRecord, 0)
	for _, file := range r.files {
		if file.RecipientID == recipientID {
			result = append(result, file)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (r *FileRepository) ListBySenderID(ctx context.Context, senderID int) ([]domain.FileRecord, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.FileRecord, 0)
	for _, file := range r.files {
		if file.SenderID == senderID {
			result = append(result, file)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}
