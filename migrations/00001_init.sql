-- +goose Up

CREATE TABLE users (
    id                    INTEGER PRIMARY KEY,
    encryption_public_key BYTEA   NOT NULL,
    signing_public_key    BYTEA   NOT NULL
);

CREATE TABLE file_records (
    id           INTEGER     PRIMARY KEY,
    size         BIGINT      NOT NULL,
    sender_id    INTEGER     NOT NULL REFERENCES users(id),
    recipient_id INTEGER     NOT NULL REFERENCES users(id),
    storage_key  TEXT        NOT NULL UNIQUE,
    file_name    TEXT        NOT NULL,
    mime_type    TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_file_records_recipient_id ON file_records(recipient_id);
CREATE INDEX idx_file_records_sender_id    ON file_records(sender_id);

CREATE TABLE sessions (
    token      TEXT        PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);

CREATE TABLE challenges (
    user_id    INTEGER     PRIMARY KEY REFERENCES users(id),
    nonce      BYTEA       NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS file_records;
DROP TABLE IF EXISTS users;
