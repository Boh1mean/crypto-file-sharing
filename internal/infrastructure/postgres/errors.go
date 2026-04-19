package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isDuplicateKey reports whether err is a PostgreSQL unique-constraint violation (23505).
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
