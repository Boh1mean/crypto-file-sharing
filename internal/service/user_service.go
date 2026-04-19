package service

import (
	"context"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) CreateUser(
	ctx context.Context,
	input domain.CreateUserInput,
) (domain.CreateUserOutput, error) {
	user := domain.User{
		ID:                  input.ID,
		EncryptionPublicKey: input.EncryptionPublicKey,
		SigningPublicKey:    input.SigningPublicKey,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return domain.CreateUserOutput{}, err
	}

	return domain.CreateUserOutput{ID: input.ID}, nil
}

func (s *UserService) GetUserPublicKeys(
	ctx context.Context,
	input domain.GetUserPublicKeysInput,
) (domain.GetUserPublicKeysOutput, error) {
	user, err := s.users.GetByID(ctx, input.ID)
	if err != nil {
		return domain.GetUserPublicKeysOutput{}, err
	}

	return domain.GetUserPublicKeysOutput{
		ID:                  user.ID,
		EncryptionPublicKey: user.EncryptionPublicKey,
		SigningPublicKey:    user.SigningPublicKey,
	}, nil
}
