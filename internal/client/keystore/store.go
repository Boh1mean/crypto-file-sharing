package keystore

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Store struct {
	path string
}

var ErrProfileNotFound = errors.New("local profile not found")

func NewDefaultStore() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	return &Store{
		path: filepath.Join(configDir, "cryptocore", "profile.json"),
	}, nil
}

func (s *Store) Save(profile Profile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, raw, 0o600)
}

func (s *Store) Load() (Profile, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Profile{}, ErrProfileNotFound
		}
		return Profile{}, err
	}

	var profile Profile
	if err := json.Unmarshal(raw, &profile); err != nil {
		return Profile{}, err
	}

	return profile, nil
}

func (s *Store) LoadIdentity() (Identity, error) {
	profile, err := s.Load()
	if err != nil {
		return Identity{}, err
	}

	encryptionPrivateKey, err := DecodeEncryptionPrivateKey(profile.EncryptionPrivateKey)
	if err != nil {
		return Identity{}, err
	}

	signingPrivateKey, err := DecodeSigningPrivateKey(profile.SigningPrivateKey)
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		ServerURL:            profile.ServerURL,
		UserID:               profile.UserID,
		EncryptionPrivateKey: encryptionPrivateKey,
		SigningPrivateKey:    signingPrivateKey,
		SessionToken:         profile.SessionToken,
	}, nil
}

func DecodeEncryptionPrivateKey(raw []byte) (*ecdh.PrivateKey, error) {
	return ecdh.P256().NewPrivateKey(raw)
}

func DecodeSigningPrivateKey(raw []byte) (*ecdsa.PrivateKey, error) {
	return x509.ParseECPrivateKey(raw)
}
