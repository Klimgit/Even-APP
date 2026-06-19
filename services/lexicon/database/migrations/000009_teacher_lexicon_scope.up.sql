ALTER TABLE lexemes
    ADD COLUMN scope TEXT NOT NULL DEFAULT 'platform'
        CHECK (scope IN ('platform', 'teacher')),
    ADD COLUMN owner_id UUID;

ALTER TABLE lexemes DROP CONSTRAINT IF EXISTS lexemes_language_id_lemma_key;

CREATE UNIQUE INDEX idx_lexemes_platform_lemma
    ON lexemes (language_id, lemma)
    WHERE scope = 'platform';

CREATE UNIQUE INDEX idx_lexemes_teacher_lemma
    ON lexemes (language_id, lemma, owner_id)
    WHERE scope = 'teacher';

CREATE INDEX idx_lexemes_teacher_owner ON lexemes (owner_id, language_id)
    WHERE scope = 'teacher';

UPDATE lexemes SET scope = 'platform' WHERE scope IS NULL;
