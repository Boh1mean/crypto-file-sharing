# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run server (listens on :8080)
go run .

# Run desktop client
go run ./cmd/desktop

# Run a single test package
go test ./internal/core/crypto/...

# Run a specific test
go test -run TestFunctionName ./internal/...

# Generate test user credentials (for manual API testing)
go run ./cmd/devkeygen -id 2
```

## Architecture

This is a Go project implementing an end-to-end encrypted file transfer system with a server and a Fyne-based desktop client.

### Key Principle

Private keys never leave the client. The server stores only public keys, encrypted containers, and metadata. All encryption/decryption happens locally.

### Layer Structure

```
internal/
├── core/           # Domain-agnostic cryptographic core
│   ├── crypto/     # Primitives: ECDH (P-256), ECDSA, AES-GCM, HKDF, SHA-256
│   └── container/  # File container protocol: BuildContainer / DecryptContainer
├── domain/         # Domain models: User, FileRecord, Session, Challenge
├── repository/     # Data access interfaces (abstractions only)
├── infrastructure/
│   └── memory/     # In-memory implementations of all repositories (dev/test)
├── service/        # Business logic: UserService, FileService, AuthService
├── transport/
│   └── httpapi/    # HTTP handlers, middleware, DTOs, router
└── client/
    ├── api/        # HTTP client for talking to the server
    ├── app/        # Client-side use cases: Register, Login, SendFile, ReceiveFile
    ├── keystore/   # Local profile + key persistence (~/.config/cryptocore/)
    └── ui/         # Fyne GUI screens
```

### Cryptographic Protocol

**File encryption** (`container.BuildContainer`):
1. Generate random file key → AES-GCM encrypt plaintext (with SHA-256 hash in AAD)
2. Generate ephemeral ECDH keypair → compute shared secret with recipient's public key
3. Derive wrap key via HKDF(shared secret, random salt) → AES-GCM wrap the file key
4. ECDSA-sign the complete container (excluding signature field)

**File decryption** (`container.DecryptContainer`):
1. Verify ECDSA signature against sender's public key
2. Recompute shared secret using ephemeral public key + recipient's private key
3. Derive wrap key → unwrap file key → decrypt content → verify SHA-256

**Authentication** (challenge-response):
1. Client POSTs to `/auth/challenge` → server stores a 2-minute nonce
2. Client signs nonce with ECDSA private key, POSTs to `/auth/verify`
3. Server issues a 30-day session token used as `Authorization` header on protected routes

### HTTP API

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/users` | No | Register user + public keys |
| POST | `/auth/challenge` | No | Get nonce for signing |
| POST | `/auth/verify` | No | Submit signature, get session token |
| GET | `/users/{id}/public-keys` | Yes | Fetch recipient public keys |
| POST | `/files` | Yes | Upload encrypted container |
| GET | `/files/{id}` | Yes | Download encrypted container |

### Storage

All repositories are currently in-memory (no database). Data is lost on server restart. The `internal/infrastructure/memory/contracts.go` file verifies interface compliance at compile time.

### Client Profile

The desktop client stores its profile (server URL, user ID, ECDH/ECDSA keys, session token) as JSON in `~/.config/cryptocore/profile.json`. Keys are stored unencrypted in plaintext.
