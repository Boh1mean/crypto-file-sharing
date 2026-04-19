package repository

import (
	"context"
	"cryptocore/internal/domain"
)

type FileRepository interface {
	Create(ctx context.Context, file domain.FileRecord) error
	GetByID(ctx context.Context, id int) (domain.FileRecord, error)
	Update(ctx context.Context, file domain.FileRecord) error
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)
	List(ctx context.Context) ([]domain.FileRecord, error)
	ListByRecipientID(ctx context.Context, recipientID int) ([]domain.FileRecord, error)
	ListBySenderID(ctx context.Context, senderID int) ([]domain.FileRecord, error)
}
