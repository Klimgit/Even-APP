ALTER TABLE user_block_progress
    ADD COLUMN IF NOT EXISTS time_spent_seconds INT NOT NULL DEFAULT 0;
