package repository

import (
	"context"

	"cryptocore/internal/domain"
)

type ChallengeRepository interface {
	Save(ctx context.Context, challenge domain.Challenge) error
	FindByUserID(ctx context.Context, userID int) (domain.Challenge, error)
	Delete(ctx context.Context, userID int) error
}
