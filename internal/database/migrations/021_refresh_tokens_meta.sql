ALTER TABLE refresh_tokens
    ADD COLUMN IF NOT EXISTS ip_address TEXT,
    ADD COLUMN IF NOT EXISTS user_agent TEXT,
    ADD COLUMN IF NOT EXISTS device_name TEXT,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ;

UPDATE refresh_tokens
SET last_used_at = created_at
WHERE last_used_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_active
    ON refresh_tokens(user_id, revoked_at, expires_at);
