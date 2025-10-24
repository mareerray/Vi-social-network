-- Add online_status column to users for tracking presence (0 = offline, 1 = online)
ALTER TABLE users ADD COLUMN online_status INTEGER DEFAULT 0;
