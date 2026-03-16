-- 010_message_reactions.sql
-- Emoji reactions for room messages and direct messages

CREATE TABLE message_reactions (
    id           BIGSERIAL   PRIMARY KEY,
    message_id   BIGINT      NOT NULL,
    message_type VARCHAR(8)  NOT NULL DEFAULT 'room', -- 'room' | 'dm'
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji        VARCHAR(64) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT message_reactions_unique UNIQUE (message_id, message_type, user_id, emoji),
    CONSTRAINT message_reactions_type_check CHECK (message_type IN ('room', 'dm'))
);

CREATE INDEX idx_reactions_msg ON message_reactions(message_id, message_type);
CREATE INDEX idx_reactions_user ON message_reactions(user_id);
