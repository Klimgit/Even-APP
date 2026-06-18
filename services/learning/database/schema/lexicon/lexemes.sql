CREATE TABLE IF NOT EXISTS lexemes (
    id UUID PRIMARY KEY,
    language_id UUID NOT NULL REFERENCES languages(id),
    lemma TEXT NOT NULL,
    part_of_speech TEXT,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
