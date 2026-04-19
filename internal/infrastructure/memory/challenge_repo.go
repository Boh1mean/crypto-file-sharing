package memory

import (
	"context"
	"sync"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type ChallengeRepository struct {
	mu         sync.RWMutex
	challenges map[int]domain.Challenge
}

func NewChallengeRepository() *ChallengeRepository {
	return &ChallengeRepository{
		challenges: make(map[int]domain.Challenge),
	}
}

func (r *ChallengeRepository) Save(ctx context.Context, challenge domain.Challenge) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.challenges[challenge.UserID] = challenge
	return nil
}

func (r *ChallengeRepository) FindByUserID(ctx context.Context, userID int) (domain.Challenge, error) {
	select {
	case <-ctx.Done():
		return domain.Challenge{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	challenge, exists := r.challenges[userID]
	if !exists {
		return domain.Challenge{}, repository.ErrChallengeNotFound
	}

	return challenge, nil
}

func (r *ChallengeRepository) Delete(ctx context.Context, userID int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.challenges[userID]; !exists {
		return repository.ErrChallengeNotFound
	}

	delete(r.challenges, userID)
	return nil
}
