CREATE TABLE course_modules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id  UUID NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_course_modules_course ON course_modules (course_id, sort_order);

ALTER TABLE lessons ADD COLUMN module_id UUID REFERENCES course_modules (id) ON DELETE CASCADE;

INSERT INTO course_modules (course_id, title, sort_order)
SELECT c.id, 'Основной', 0
FROM courses c
WHERE EXISTS (SELECT 1 FROM lessons l WHERE l.course_id = c.id)
   OR NOT EXISTS (SELECT 1 FROM lessons l2 WHERE l2.course_id = c.id);

UPDATE lessons l
SET module_id = m.id
FROM course_modules m
WHERE m.course_id = l.course_id AND l.module_id IS NULL;

ALTER TABLE lessons ALTER COLUMN module_id SET NOT NULL;

CREATE INDEX idx_lessons_module_sort ON lessons (module_id, sort_order);
