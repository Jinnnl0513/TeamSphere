CREATE TABLE IF NOT EXISTS room_role_permissions (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    can_send BOOLEAN NOT NULL DEFAULT TRUE,
    can_upload BOOLEAN NOT NULL DEFAULT TRUE,
    can_pin BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_members BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_settings BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_messages BOOLEAN NOT NULL DEFAULT FALSE,
    can_mention_all BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(room_id, role)
);

CREATE INDEX IF NOT EXISTS idx_room_role_permissions_room_role ON room_role_permissions(room_id, role);
