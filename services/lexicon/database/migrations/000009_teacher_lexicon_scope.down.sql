DROP INDEX IF EXISTS idx_lexemes_teacher_owner;
DROP INDEX IF EXISTS idx_lexemes_teacher_lemma;
DROP INDEX IF EXISTS idx_lexemes_platform_lemma;

ALTER TABLE lexemes DROP COLUMN IF EXISTS owner_id;
ALTER TABLE lexemes DROP COLUMN IF EXISTS scope;

ALTER TABLE lexemes ADD CONSTRAINT lexemes_language_id_lemma_key UNIQUE (language_id, lemma);
