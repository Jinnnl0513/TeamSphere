-- 007_refresh_tokens.sql: Refresh token support for secure authentication

-- Refresh tokens table
-- Stores long-lived refresh tokens that can be revoked
CREATE TABLE refresh_tokens (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(128) NOT NULL UNIQUE,  -- SHA-256 hash of the token
    expires_at  TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    revoked_at  TIMESTAMPTZ  DEFAULT NULL      -- NULL means active, non-NULL means revoked
);

-- Index for quick lookup by user
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id, expires_at);
-- Index for cleanup of expired tokens
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
