ALTER TABLE chats ADD COLUMN last_synced INTEGER DEFAULT 0;
UPDATE chats SET last_synced = 1;
