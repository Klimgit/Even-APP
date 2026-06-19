CREATE TABLE IF NOT EXISTS media_assets (
    id UUID PRIMARY KEY,
    scope TEXT NOT NULL,
    language_id UUID NOT NULL,
    owner_id UUID,
    object_key TEXT NOT NULL,
    bucket TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    media_kind TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    display_name TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
