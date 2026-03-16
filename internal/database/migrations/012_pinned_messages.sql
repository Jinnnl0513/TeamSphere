-- 012_pinned_messages.sql: Add pinned messages for rooms

CREATE TABLE pinned_messages (
    id         BIGSERIAL   PRIMARY KEY,
    room_id    BIGINT      NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    msg_id     BIGINT      NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    pinned_by  BIGINT      NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_pinned_messages_room_msg ON pinned_messages(room_id, msg_id);
CREATE INDEX idx_pinned_messages_room_created ON pinned_messages(room_id, created_at DESC);
