CREATE TABLE IF NOT EXISTS room_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    actor_id BIGINT REFERENCES users(id),
    action TEXT NOT NULL,
    before JSONB,
    after JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_room_audit_logs_room_created ON room_audit_logs(room_id, created_at);
