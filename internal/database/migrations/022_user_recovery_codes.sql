CREATE TABLE IF NOT EXISTS user_recovery_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(128) NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_recovery_user_code ON user_recovery_codes(user_id, code_hash);
CREATE INDEX IF NOT EXISTS ix_recovery_user_used ON user_recovery_codes(user_id, used_at);
