-- 001_init.sql: Initial schema for TeamSphere

-- Users
CREATE TABLE users (
    id         BIGSERIAL    PRIMARY KEY,
    username   VARCHAR(32)  NOT NULL UNIQUE,
    password   VARCHAR(128) NOT NULL,
    nickname   VARCHAR(32)  NOT NULL DEFAULT '',
    avatar_url VARCHAR(512) NOT NULL DEFAULT '',
    role       VARCHAR(16)  NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Rooms
CREATE TABLE rooms (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    description VARCHAR(256) NOT NULL DEFAULT '',
    creator_id  BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Messages
CREATE TABLE messages (
    id            BIGSERIAL    PRIMARY KEY,
    content       TEXT         NOT NULL,
    user_id       BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    room_id       BIGINT       NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    msg_type      VARCHAR(16)  NOT NULL DEFAULT 'text',
    mentions      BIGINT[]     NOT NULL DEFAULT '{}',
    deleted_at    TIMESTAMPTZ,
    client_msg_id VARCHAR(64)  UNIQUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Room members
CREATE TABLE room_members (
    room_id     BIGINT      NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(16) NOT NULL DEFAULT 'member',
    muted_until TIMESTAMPTZ,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (room_id, user_id)
);

-- Room invites
CREATE TABLE room_invites (
    id         BIGSERIAL   PRIMARY KEY,
    room_id    BIGINT      NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    inviter_id BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_room_invites_pending ON room_invites(room_id, invitee_id) WHERE status = 'pending';

-- Friendships
CREATE TABLE friendships (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id  BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id != friend_id)
);
CREATE UNIQUE INDEX idx_friendships_pair ON friendships(
    LEAST(user_id, friend_id),
    GREATEST(user_id, friend_id)
);

-- Direct messages
CREATE TABLE direct_messages (
    id            BIGSERIAL    PRIMARY KEY,
    content       TEXT         NOT NULL,
    sender_id     BIGINT       NOT NULL,
    receiver_id   BIGINT       NOT NULL,
    msg_type      VARCHAR(16)  NOT NULL DEFAULT 'text',
    deleted_at    TIMESTAMPTZ,
    client_msg_id VARCHAR(64)  UNIQUE,
    read_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Token blacklist
CREATE TABLE token_blacklist (
    id         BIGSERIAL    PRIMARY KEY,
    token_jti  VARCHAR(64)  NOT NULL UNIQUE,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at);

-- System settings
CREATE TABLE system_settings (
    key        VARCHAR(64)  PRIMARY KEY,
    value      TEXT         NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_messages_room_id_id   ON messages(room_id, id DESC);
CREATE INDEX idx_messages_user_id      ON messages(user_id);
CREATE INDEX idx_room_members_user     ON room_members(user_id);
CREATE INDEX idx_room_invites_invitee  ON room_invites(invitee_id, status);
CREATE INDEX idx_friendships_user      ON friendships(user_id, status);
CREATE INDEX idx_friendships_friend    ON friendships(friend_id, status);
CREATE INDEX idx_dm_pair               ON direct_messages(
    LEAST(sender_id, receiver_id),
    GREATEST(sender_id, receiver_id),
    id DESC
);
CREATE INDEX idx_dm_sender             ON direct_messages(sender_id, id DESC);
CREATE INDEX idx_dm_receiver           ON direct_messages(receiver_id, id DESC);
