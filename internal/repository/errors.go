package repository

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrFileNotFound      = errors.New("file not found")
	ErrFileAlreadyExists = errors.New("file already exists")

	ErrContainerNotFound = errors.New("container not found")

	ErrSessionNotFound   = errors.New("session not found")
	ErrChallengeNotFound = errors.New("challenge not found")
)
