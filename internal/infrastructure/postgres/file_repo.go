package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type FileRepository struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepository {
	return &FileRepository{pool: pool}
}

func (r *FileRepository) Create(ctx context.Context, file domain.FileRecord) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO file_records (size, sender_id, recipient_id, storage_key, file_name, mime_type, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		file.Size, file.SenderID, file.RecipientID,
		file.StorageKey, file.FileName, file.MimeType, file.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int) (domain.FileRecord, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, size, sender_id, recipient_id, storage_key, file_name, mime_type, created_at
		 FROM file_records WHERE id = $1`,
		id,
	)

	var f domain.FileRecord
	if err := row.Scan(&f.ID, &f.Size, &f.SenderID, &f.RecipientID,
		&f.StorageKey, &f.FileName, &f.MimeType, &f.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FileRecord{}, repository.ErrFileNotFound
		}
		return domain.FileRecord{}, err
	}

	return f, nil
}

func (r *FileRepository) Update(ctx context.Context, file domain.FileRecord) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE file_records
		 SET size=$2, sender_id=$3, recipient_id=$4, storage_key=$5, file_name=$6, mime_type=$7, created_at=$8
		 WHERE id=$1`,
		file.ID, file.Size, file.SenderID, file.RecipientID,
		file.StorageKey, file.FileName, file.MimeType, file.CreatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrFileNotFound
	}
	return nil
}

func (r *FileRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM file_records WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrFileNotFound
	}
	return nil
}

func (r *FileRepository) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM file_records WHERE id = $1)`, id,
	).Scan(&exists)
	return exists, err
}

func (r *FileRepository) List(ctx context.Context) ([]domain.FileRecord, error) {
	return r.queryFiles(ctx,
		`SELECT id, size, sender_id, recipient_id, storage_key, file_name, mime_type, created_at
		 FROM file_records ORDER BY id`,
	)
}

func (r *FileRepository) ListByRecipientID(ctx context.Context, recipientID int) ([]domain.FileRecord, error) {
	return r.queryFiles(ctx,
		`SELECT id, size, sender_id, recipient_id, storage_key, file_name, mime_type, created_at
		 FROM file_records WHERE recipient_id = $1 ORDER BY id`,
		recipientID,
	)
}

func (r *FileRepository) ListBySenderID(ctx context.Context, senderID int) ([]domain.FileRecord, error) {
	return r.queryFiles(ctx,
		`SELECT id, size, sender_id, recipient_id, storage_key, file_name, mime_type, created_at
		 FROM file_records WHERE sender_id = $1 ORDER BY id`,
		senderID,
	)
}

func (r *FileRepository) queryFiles(ctx context.Context, query string, args ...any) ([]domain.FileRecord, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []domain.FileRecord
	for rows.Next() {
		var f domain.FileRecord
		if err := rows.Scan(&f.ID, &f.Size, &f.SenderID, &f.RecipientID,
			&f.StorageKey, &f.FileName, &f.MimeType, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, rows.Err()
}
