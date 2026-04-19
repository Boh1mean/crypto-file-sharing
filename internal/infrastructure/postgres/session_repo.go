package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Save(ctx context.Context, session domain.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at, created_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (token) DO UPDATE
		   SET user_id = EXCLUDED.user_id,
		       expires_at = EXCLUDED.expires_at,
		       created_at = EXCLUDED.created_at`,
		session.Token, session.UserID, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (domain.Session, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT token, user_id, expires_at, created_at FROM sessions WHERE token = $1`,
		token,
	)

	var s domain.Session
	if err := row.Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, repository.ErrSessionNotFound
		}
		return domain.Session{}, err
	}

	return s, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}
