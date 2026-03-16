CREATE TABLE IF NOT EXISTS message_reads (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    last_read_msg_id BIGINT NOT NULL DEFAULT 0,
    read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, room_id)
);

CREATE INDEX IF NOT EXISTS idx_message_reads_room_user ON message_reads(room_id, user_id);
