CREATE TABLE IF NOT EXISTS languages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    native_name TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'ltr' CHECK (direction IN ('ltr', 'rtl')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope TEXT NOT NULL DEFAULT 'platform' CHECK (scope IN ('platform', 'teacher')),
    language_id UUID NOT NULL REFERENCES languages(id),
    owner_id UUID,
    object_key TEXT UNIQUE NOT NULL,
    bucket TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    media_kind TEXT NOT NULL CHECK (media_kind IN ('image', 'audio', 'video')),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    width INT,
    height INT,
    duration_ms INT,
    in_library BOOLEAN NOT NULL DEFAULT false,
    display_name TEXT,
    linked_lexeme_id UUID,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (NOT in_library OR (display_name IS NOT NULL AND char_length(display_name) BETWEEN 1 AND 120))
);

CREATE INDEX IF NOT EXISTS idx_media_platform_library
    ON media_assets (language_id, in_library)
    WHERE scope = 'platform' AND in_library;
