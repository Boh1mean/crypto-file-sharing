package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type ChallengeRepository struct {
	pool *pgxpool.Pool
}

func NewChallengeRepository(pool *pgxpool.Pool) *ChallengeRepository {
	return &ChallengeRepository{pool: pool}
}

func (r *ChallengeRepository) Save(ctx context.Context, challenge domain.Challenge) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO challenges (user_id, nonce, expires_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE
		   SET nonce = EXCLUDED.nonce,
		       expires_at = EXCLUDED.expires_at`,
		challenge.UserID, challenge.Nonce, challenge.ExpiresAt,
	)
	return err
}

func (r *ChallengeRepository) FindByUserID(ctx context.Context, userID int) (domain.Challenge, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT user_id, nonce, expires_at FROM challenges WHERE user_id = $1`,
		userID,
	)

	var c domain.Challenge
	if err := row.Scan(&c.UserID, &c.Nonce, &c.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Challenge{}, repository.ErrChallengeNotFound
		}
		return domain.Challenge{}, err
	}

	return c, nil
}

func (r *ChallengeRepository) Delete(ctx context.Context, userID int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM challenges WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrChallengeNotFound
	}
	return nil
}
