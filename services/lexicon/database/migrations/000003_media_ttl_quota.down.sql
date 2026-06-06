ALTER TABLE media_assets ADD COLUMN IF NOT EXISTS in_library BOOLEAN NOT NULL DEFAULT true;

UPDATE media_assets SET in_library = true;

ALTER TABLE media_assets DROP COLUMN IF EXISTS expires_at;

DROP INDEX IF EXISTS idx_media_platform_active;
DROP INDEX IF EXISTS idx_media_uploaded_by;

CREATE INDEX IF NOT EXISTS idx_media_platform_library
    ON media_assets (language_id, in_library)
    WHERE scope = 'platform' AND in_library;
