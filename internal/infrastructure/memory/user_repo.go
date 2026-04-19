package memory

import (
	"context"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
	"sort"
	"sync"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[int]domain.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int]domain.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return repository.ErrUserAlreadyExists
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (domain.User, error) {
	select {
	case <-ctx.Done():
		return domain.User{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return domain.User{}, repository.ErrUserNotFound
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return repository.ErrUserNotFound
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return repository.ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

func (r *UserRepository) Exists(ctx context.Context, id int) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.users[id]
	return exists, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.User, 0, len(r.users))
	for _, user := range r.users {
		result = append(result, user)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}
