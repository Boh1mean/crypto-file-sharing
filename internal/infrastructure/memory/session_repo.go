package memory

import (
	"context"
	"sync"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]domain.Session),
	}
}

func (r *SessionRepository) Save(ctx context.Context, session domain.Session) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[session.Token] = session
	return nil
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (domain.Session, error) {
	select {
	case <-ctx.Done():
		return domain.Session{}, ctx.Err()
	default:
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[token]
	if !exists {
		return domain.Session{}, repository.ErrSessionNotFound
	}

	return session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[token]; !exists {
		return repository.ErrSessionNotFound
	}

	delete(r.sessions, token)
	return nil
}
