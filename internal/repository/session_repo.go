package repository

import (
	"context"

	"cryptocore/internal/domain"
)

type SessionRepository interface {
	Save(ctx context.Context, session domain.Session) error
	FindByToken(ctx context.Context, token string) (domain.Session, error)
	Delete(ctx context.Context, token string) error
}
