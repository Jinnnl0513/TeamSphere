-- 013_message_search_fts.sql: Add FTS index for message search

CREATE INDEX idx_messages_content_fts ON messages USING GIN (to_tsvector('simple', content));
