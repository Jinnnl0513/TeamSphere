-- 004_message_reply.sql
-- Add reply_to_id to messages and direct_messages tables

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS reply_to_id BIGINT REFERENCES messages(id) ON DELETE SET NULL;

ALTER TABLE direct_messages
    ADD COLUMN IF NOT EXISTS reply_to_id BIGINT REFERENCES direct_messages(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_messages_reply ON messages(reply_to_id) WHERE reply_to_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_dm_reply ON direct_messages(reply_to_id) WHERE reply_to_id IS NOT NULL;
