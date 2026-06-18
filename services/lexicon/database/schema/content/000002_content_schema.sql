-- Read-only schema mirror for sqlc content queries (matches content service migrations).

CREATE TABLE courses (
    id                 UUID PRIMARY KEY,
    title              TEXT NOT NULL,
    target_language_id UUID NOT NULL,
    ui_language_id     UUID NOT NULL,
    owner_id           UUID NOT NULL,
    is_published       BOOLEAN NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL
);

CREATE TABLE lessons (
    id           UUID PRIMARY KEY,
    course_id    UUID NOT NULL,
    title        TEXT NOT NULL,
    sort_order   INT NOT NULL,
    version      INT NOT NULL,
    status       TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE lesson_blocks (
    id            UUID PRIMARY KEY,
    lesson_id     UUID NOT NULL,
    section_id    UUID,
    sort_order    INT NOT NULL,
    display_label TEXT,
    title         TEXT,
    block_type    TEXT NOT NULL,
    config        JSONB NOT NULL,
    is_homework   BOOLEAN NOT NULL
);
