CREATE TABLE IF NOT EXISTS lexeme_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_lexeme_id UUID NOT NULL REFERENCES lexemes(id) ON DELETE CASCADE,
    target_language_id UUID NOT NULL REFERENCES languages(id),
    text TEXT NOT NULL,
    target_lexeme_id UUID REFERENCES lexemes(id)
);
