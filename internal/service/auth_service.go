package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"cryptocore/internal/core/crypto"
	"cryptocore/internal/domain"
	"cryptocore/internal/repository"
)

const (
	challengeTTL   = 2 * time.Minute
	sessionTTL     = 30 * 24 * time.Hour
	nonceSize      = 32
	sessionKeySize = 32
)

var (
	ErrChallengeExpired  = errors.New("challenge expired or not found")
	ErrInvalidSignature  = errors.New("invalid challenge signature")
	ErrSessionExpired    = errors.New("session expired")
	ErrSessionNotFound   = errors.New("session not found")
)

type AuthService struct {
	users      repository.UserRepository
	sessions   repository.SessionRepository
	challenges repository.ChallengeRepository
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	challenges repository.ChallengeRepository,
) *AuthService {
	return &AuthService{
		users:      users,
		sessions:   sessions,
		challenges: challenges,
	}
}

// CreateChallenge генерирует случайный nonce для пользователя и сохраняет его с TTL 2 минуты.
func (s *AuthService) CreateChallenge(ctx context.Context, userID int) ([]byte, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	nonce, err := crypto.RandomBytes(nonceSize)
	if err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	challenge := domain.Challenge{
		Nonce:     nonce,
		UserID:    userID,
		ExpiresAt: time.Now().Add(challengeTTL),
	}

	if err := s.challenges.Save(ctx, challenge); err != nil {
		return nil, fmt.Errorf("save challenge: %w", err)
	}

	return nonce, nil
}

// VerifyChallenge проверяет подпись nonce ECDSA ключом пользователя.
// При успехе удаляет challenge и выдаёт долгоживущий session token.
func (s *AuthService) VerifyChallenge(ctx context.Context, userID int, signature []byte) (domain.Session, error) {
	challenge, err := s.challenges.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrChallengeNotFound) {
			return domain.Session{}, ErrChallengeExpired
		}
		return domain.Session{}, err
	}

	if challenge.IsExpired() {
		_ = s.challenges.Delete(ctx, userID)
		return domain.Session{}, ErrChallengeExpired
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.Session{}, fmt.Errorf("get user: %w", err)
	}

	if !crypto.Verify(user.SigningPublicKey, challenge.Nonce, signature) {
		return domain.Session{}, ErrInvalidSignature
	}

	// Challenge одноразовый — удаляем после успешной проверки.
	_ = s.challenges.Delete(ctx, userID)

	tokenBytes, err := crypto.RandomBytes(sessionKeySize)
	if err != nil {
		return domain.Session{}, fmt.Errorf("generate session token: %w", err)
	}

	session := domain.Session{
		Token:     hex.EncodeToString(tokenBytes),
		UserID:    userID,
		ExpiresAt: time.Now().Add(sessionTTL),
		CreatedAt: time.Now(),
	}

	if err := s.sessions.Save(ctx, session); err != nil {
		return domain.Session{}, fmt.Errorf("save session: %w", err)
	}

	return session, nil
}

// ValidateSession проверяет токен и возвращает userID. Используется в middleware.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (int, error) {
	session, err := s.sessions.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return 0, ErrSessionNotFound
		}
		return 0, err
	}

	if session.IsExpired() {
		_ = s.sessions.Delete(ctx, token)
		return 0, ErrSessionExpired
	}

	return session.UserID, nil
}

// Logout инвалидирует сессию пользователя.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	err := s.sessions.Delete(ctx, token)
	if errors.Is(err, repository.ErrSessionNotFound) {
		return nil
	}
	return err
}
