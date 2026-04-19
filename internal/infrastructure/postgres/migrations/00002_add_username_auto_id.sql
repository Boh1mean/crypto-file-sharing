-- +goose Up

-- Make users.id auto-generated
CREATE SEQUENCE IF NOT EXISTS users_id_seq;
ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq');
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 0) + 1, false);

-- Add username column
ALTER TABLE users ADD COLUMN username TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
ALTER TABLE users ADD CONSTRAINT users_username_length CHECK (char_length(username) <= 32);
ALTER TABLE users ALTER COLUMN username DROP DEFAULT;

-- Make file_records.id auto-generated
CREATE SEQUENCE IF NOT EXISTS file_records_id_seq;
ALTER TABLE file_records ALTER COLUMN id SET DEFAULT nextval('file_records_id_seq');
SELECT setval('file_records_id_seq', COALESCE((SELECT MAX(id) FROM file_records), 0) + 1, false);

-- +goose Down

ALTER TABLE file_records ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS file_records_id_seq;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_length;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE users DROP COLUMN IF EXISTS username;
ALTER TABLE users ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS users_id_seq;
