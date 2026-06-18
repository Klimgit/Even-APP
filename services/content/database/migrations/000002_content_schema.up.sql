CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    target_language_id UUID NOT NULL,
    ui_language_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_courses_owner ON courses (owner_id);

CREATE TABLE course_invite_codes (
    course_id UUID PRIMARY KEY REFERENCES courses (id) ON DELETE CASCADE,
    code CHAR(8) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    published_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lessons_course_status ON lessons (course_id, status);
CREATE INDEX idx_lessons_course_sort ON lessons (course_id, sort_order);

CREATE TABLE lesson_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    section_kind TEXT NOT NULL DEFAULT 'content' CHECK (section_kind IN ('content', 'homework', 'results'))
);

CREATE INDEX idx_lesson_sections_lesson ON lesson_sections (lesson_id, sort_order);

CREATE TABLE lesson_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    section_id UUID REFERENCES lesson_sections (id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0,
    display_label TEXT,
    title TEXT,
    block_type TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    is_homework BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_lesson_blocks_lesson ON lesson_blocks (lesson_id, sort_order);
