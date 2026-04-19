package repository

import (
	"context"
	"cryptocore/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id int) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)
	List(ctx context.Context) ([]domain.User, error)
}
