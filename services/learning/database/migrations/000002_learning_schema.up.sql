CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE course_enrollments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    course_id       UUID NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    enrolled_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    enrolled_by     UUID,
    UNIQUE (user_id, course_id)
);

CREATE INDEX idx_course_enrollments_user ON course_enrollments (user_id);
CREATE INDEX idx_course_enrollments_course ON course_enrollments (course_id);

CREATE TABLE user_block_progress (
    user_id           UUID NOT NULL,
    lesson_block_id   UUID NOT NULL,
    status            TEXT NOT NULL DEFAULT 'not_started',
    score             REAL NOT NULL DEFAULT 0,
    attempts          INT NOT NULL DEFAULT 0,
    last_attempt_at   TIMESTAMPTZ,
    PRIMARY KEY (user_id, lesson_block_id)
);

CREATE TABLE block_attempts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL,
    lesson_block_id   UUID NOT NULL,
    sub_item_index    INT NOT NULL DEFAULT 0,
    is_correct        BOOLEAN NOT NULL,
    response          JSONB NOT NULL,
    context           TEXT NOT NULL DEFAULT 'lesson',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_block_attempts_user_block ON block_attempts (user_id, lesson_block_id);

CREATE TABLE user_review_items (
    user_id           UUID NOT NULL,
    lesson_block_id   UUID NOT NULL,
    sub_item_index    INT NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'pending',
    failure_count     INT NOT NULL DEFAULT 1,
    consecutive_ok    INT NOT NULL DEFAULT 0,
    first_failed_at   TIMESTAMPTZ,
    last_failed_at    TIMESTAMPTZ,
    last_attempt_at   TIMESTAMPTZ,
    due_at            TIMESTAMPTZ NOT NULL,
    source_lesson_id  UUID NOT NULL,
    course_id         UUID NOT NULL,
    PRIMARY KEY (user_id, lesson_block_id, sub_item_index)
);

CREATE INDEX idx_review_due ON user_review_items (user_id, status, due_at);

CREATE TABLE user_vocabulary (
    user_id         UUID NOT NULL,
    lexeme_id       UUID NOT NULL,
    course_id       UUID,
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    mastery         REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, lexeme_id)
);

CREATE INDEX idx_user_vocabulary_course ON user_vocabulary (user_id, course_id);

CREATE TABLE published_lesson_snapshots (
    lesson_id     UUID PRIMARY KEY,
    course_id     UUID NOT NULL,
    version       INT NOT NULL,
    snapshot      JSONB NOT NULL,
    published_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_published_lessons_course ON published_lesson_snapshots (course_id);
