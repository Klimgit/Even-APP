-- Read-only schema mirror for sqlc content queries (matches content service migrations).

CREATE TABLE courses (
    id                 UUID PRIMARY KEY,
    title              TEXT NOT NULL,
    target_language_id UUID NOT NULL,
    ui_language_id     UUID NOT NULL,
    owner_id           UUID NOT NULL,
    is_published       BOOLEAN NOT NULL,
    visibility         TEXT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL
);

CREATE TABLE course_invite_codes (
    course_id  UUID PRIMARY KEY,
    code       CHAR(8) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE course_modules (
    id         UUID PRIMARY KEY,
    course_id  UUID NOT NULL,
    title      TEXT NOT NULL,
    sort_order INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE lessons (
    id           UUID PRIMARY KEY,
    course_id    UUID NOT NULL,
    module_id    UUID NOT NULL,
    title        TEXT NOT NULL,
    sort_order   INT NOT NULL,
    version      INT NOT NULL,
    status       TEXT NOT NULL,
    published_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL
);

CREATE TABLE lesson_sections (
    id           UUID PRIMARY KEY,
    lesson_id    UUID NOT NULL,
    title        TEXT NOT NULL,
    sort_order   INT NOT NULL,
    section_kind TEXT NOT NULL
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

CREATE TABLE block_lexeme_refs (
    lesson_block_id UUID NOT NULL,
    lexeme_id       UUID NOT NULL,
    form_id         UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    role            TEXT NOT NULL,
    PRIMARY KEY (lesson_block_id, lexeme_id, form_id)
);
