CREATE TABLE IF NOT EXISTS languages (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    native_name TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'ltr',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
