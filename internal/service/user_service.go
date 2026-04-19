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
		Username:            input.Username,
		EncryptionPublicKey: input.EncryptionPublicKey,
		SigningPublicKey:    input.SigningPublicKey,
	}

	id, err := s.users.Create(ctx, user)
	if err != nil {
		return domain.CreateUserOutput{}, err
	}

	return domain.CreateUserOutput{ID: id}, nil
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
		Username:            user.Username,
		EncryptionPublicKey: user.EncryptionPublicKey,
		SigningPublicKey:    user.SigningPublicKey,
	}, nil
}

func (s *UserService) GetUserByUsername(
	ctx context.Context,
	input domain.GetUserByUsernameInput,
) (domain.GetUserByUsernameOutput, error) {
	user, err := s.users.GetByUsername(ctx, input.Username)
	if err != nil {
		return domain.GetUserByUsernameOutput{}, err
	}

	return domain.GetUserByUsernameOutput{
		ID:                  user.ID,
		EncryptionPublicKey: user.EncryptionPublicKey,
		SigningPublicKey:    user.SigningPublicKey,
	}, nil
}
