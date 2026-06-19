DROP INDEX IF EXISTS idx_lessons_module_sort;
ALTER TABLE lessons DROP COLUMN IF EXISTS module_id;
DROP TABLE IF EXISTS course_modules;
