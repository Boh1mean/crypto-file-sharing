package postgres

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (int, error) {
	encKey := user.EncryptionPublicKey.Bytes()
	sigKey, err := x509.MarshalPKIXPublicKey(user.SigningPublicKey)
	if err != nil {
		return 0, err
	}

	var id int
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (encryption_public_key, signing_public_key, username)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		encKey, sigKey, user.Username,
	).Scan(&id)
	if err != nil {
		if isDuplicateKey(err) {
			return 0, repository.ErrUserAlreadyExists
		}
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, encryption_public_key, signing_public_key FROM users WHERE id = $1`,
		id,
	)
	return scanUser(row)
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, username, encryption_public_key, signing_public_key FROM users WHERE username = $1`,
		username,
	)
	return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	encKey := user.EncryptionPublicKey.Bytes()
	sigKey, err := x509.MarshalPKIXPublicKey(user.SigningPublicKey)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET encryption_public_key = $2, signing_public_key = $3, username = $4 WHERE id = $1`,
		user.ID, encKey, sigKey, user.Username,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists)
	return exists, err
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, username, encryption_public_key, signing_public_key FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (domain.User, error) {
	var u domain.User
	var encKeyBytes, sigKeyBytes []byte
	if err := s.Scan(&u.ID, &u.Username, &encKeyBytes, &sigKeyBytes); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, repository.ErrUserNotFound
		}
		return domain.User{}, err
	}

	encKey, err := ecdh.P256().NewPublicKey(encKeyBytes)
	if err != nil {
		return domain.User{}, err
	}
	sigKey, err := parseECDSAPublicKey(sigKeyBytes)
	if err != nil {
		return domain.User{}, err
	}

	u.EncryptionPublicKey = encKey
	u.SigningPublicKey = sigKey
	return u, nil
}

func parseECDSAPublicKey(b []byte) (*ecdsa.PublicKey, error) {
	parsed, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, err
	}
	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}
	return key, nil
}
