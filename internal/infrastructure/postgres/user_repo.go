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

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	encKey := user.EncryptionPublicKey.Bytes()
	sigKey, err := x509.MarshalPKIXPublicKey(user.SigningPublicKey)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO users (id, encryption_public_key, signing_public_key)
		 VALUES ($1, $2, $3)`,
		user.ID, encKey, sigKey,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return repository.ErrUserAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, encryption_public_key, signing_public_key FROM users WHERE id = $1`,
		id,
	)

	var u domain.User
	var encKeyBytes, sigKeyBytes []byte
	if err := row.Scan(&u.ID, &encKeyBytes, &sigKeyBytes); err != nil {
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

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	encKey := user.EncryptionPublicKey.Bytes()
	sigKey, err := x509.MarshalPKIXPublicKey(user.SigningPublicKey)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET encryption_public_key = $2, signing_public_key = $3 WHERE id = $1`,
		user.ID, encKey, sigKey,
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
		`SELECT id, encryption_public_key, signing_public_key FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var encKeyBytes, sigKeyBytes []byte
		if err := rows.Scan(&u.ID, &encKeyBytes, &sigKeyBytes); err != nil {
			return nil, err
		}

		encKey, err := ecdh.P256().NewPublicKey(encKeyBytes)
		if err != nil {
			return nil, err
		}
		sigKey, err := parseECDSAPublicKey(sigKeyBytes)
		if err != nil {
			return nil, err
		}

		u.EncryptionPublicKey = encKey
		u.SigningPublicKey = sigKey
		users = append(users, u)
	}

	return users, rows.Err()
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
