CREATE TABLE IF NOT EXISTS oauth_identities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_oauth_provider_subject ON oauth_identities(provider, subject);
CREATE UNIQUE INDEX IF NOT EXISTS ux_oauth_user_provider ON oauth_identities(user_id, provider);
CREATE INDEX IF NOT EXISTS ix_oauth_user_id ON oauth_identities(user_id);
