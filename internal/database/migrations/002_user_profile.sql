-- Add bio and profile_color fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS bio          TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS profile_color VARCHAR(7)  NOT NULL DEFAULT '#6c5dd3';

-- Add deleted_at soft-delete field for account deletion
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;
