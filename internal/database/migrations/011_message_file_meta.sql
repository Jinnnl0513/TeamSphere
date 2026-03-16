-- 011_message_file_meta.sql: Add file metadata for messages and direct messages

ALTER TABLE messages
    ADD COLUMN file_size BIGINT,
    ADD COLUMN mime_type VARCHAR(128);

ALTER TABLE direct_messages
    ADD COLUMN file_size BIGINT,
    ADD COLUMN mime_type VARCHAR(128);
