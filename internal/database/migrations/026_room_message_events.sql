CREATE TABLE IF NOT EXISTS room_message_events (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id),
    event_type TEXT NOT NULL,
    meta JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_room_message_events_room_created ON room_message_events(room_id, created_at);
