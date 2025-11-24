-- +migrate Up
ALTER TABLE users
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN role VARCHAR(50) NOT NULL DEFAULT 'user';

-- +migrate Down
ALTER TABLE users
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS role;
