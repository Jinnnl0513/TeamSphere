-- Adds forward metadata for room and direct messages
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS forward_meta JSONB;

ALTER TABLE direct_messages
    ADD COLUMN IF NOT EXISTS forward_meta JSONB;
