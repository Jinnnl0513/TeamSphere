-- 005_invite_links.sql
-- Invite link system with short codes, expiry and usage limits

CREATE TABLE IF NOT EXISTS invite_links (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(16) NOT NULL UNIQUE,         -- random 8-char URL-safe code
    room_id     BIGINT      NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    creator_id  BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_uses    INT         NOT NULL DEFAULT 0,      -- 0 = unlimited
    uses        INT         NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ,                         -- NULL = never expires
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_links_code    ON invite_links(code);
CREATE INDEX IF NOT EXISTS idx_invite_links_room    ON invite_links(room_id);
