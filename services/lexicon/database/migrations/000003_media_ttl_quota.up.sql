UPDATE media_assets SET display_name = '(uploading)' WHERE display_name IS NULL;

ALTER TABLE media_assets
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

ALTER TABLE media_assets
    ALTER COLUMN display_name SET NOT NULL;

ALTER TABLE media_assets
    DROP CONSTRAINT IF EXISTS media_assets_check;

ALTER TABLE media_assets
    ADD CONSTRAINT media_assets_display_name_len
    CHECK (char_length(display_name) BETWEEN 1 AND 120);

ALTER TABLE media_assets DROP COLUMN IF EXISTS in_library;

DROP INDEX IF EXISTS idx_media_platform_library;

CREATE INDEX IF NOT EXISTS idx_media_platform_active
    ON media_assets (language_id, created_at DESC)
    WHERE scope = 'platform';

CREATE INDEX IF NOT EXISTS idx_media_uploaded_by
    ON media_assets (uploaded_by);
