CREATE TABLE IF NOT EXISTS room_join_requests (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    reason TEXT,
    reviewer_id BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(room_id, user_id, status)
);

CREATE INDEX IF NOT EXISTS idx_room_join_requests_room_status ON room_join_requests(room_id, status);
