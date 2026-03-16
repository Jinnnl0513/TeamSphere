-- 003_email_verification.sql: Add email verification support

-- Add email fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email             VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email_verified_at  TIMESTAMPTZ  DEFAULT NULL;

-- Email verification codes table
CREATE TABLE IF NOT EXISTS email_verifications (
    id         BIGSERIAL    PRIMARY KEY,
    email      VARCHAR(255) NOT NULL,
    code       VARCHAR(6)   NOT NULL,
    attempts   INT          NOT NULL DEFAULT 0,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_verifications_email ON email_verifications(email, used, expires_at);
