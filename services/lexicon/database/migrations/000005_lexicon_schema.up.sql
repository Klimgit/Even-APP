CREATE TABLE IF NOT EXISTS alphabet_letters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    character TEXT NOT NULL,
    upper_char TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    label TEXT,
    UNIQUE (language_id, character)
);

CREATE INDEX IF NOT EXISTS idx_alphabet_letters_lang_sort
    ON alphabet_letters (language_id, sort_order);

CREATE TABLE IF NOT EXISTS sounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    ipa TEXT,
    description TEXT,
    audio_key TEXT
);

CREATE INDEX IF NOT EXISTS idx_sounds_language ON sounds (language_id);

CREATE TABLE IF NOT EXISTS letter_sounds (
    letter_id UUID NOT NULL REFERENCES alphabet_letters(id) ON DELETE CASCADE,
    sound_id UUID NOT NULL REFERENCES sounds(id) ON DELETE CASCADE,
    PRIMARY KEY (letter_id, sound_id)
);

CREATE TABLE IF NOT EXISTS lexemes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    lemma TEXT NOT NULL,
    part_of_speech TEXT,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (language_id, lemma)
);

CREATE INDEX IF NOT EXISTS idx_lexemes_language ON lexemes (language_id);

CREATE TABLE IF NOT EXISTS lexeme_forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lexeme_id UUID NOT NULL REFERENCES lexemes(id) ON DELETE CASCADE,
    form TEXT NOT NULL,
    tags JSONB NOT NULL DEFAULT '{}',
    UNIQUE (lexeme_id, form)
);

CREATE TABLE IF NOT EXISTS lexeme_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_lexeme_id UUID NOT NULL REFERENCES lexemes(id) ON DELETE CASCADE,
    target_language_id UUID NOT NULL REFERENCES languages(id),
    text TEXT NOT NULL,
    target_lexeme_id UUID REFERENCES lexemes(id)
);

CREATE INDEX IF NOT EXISTS idx_lexeme_translations_source
    ON lexeme_translations (source_lexeme_id);

CREATE TABLE IF NOT EXISTS lexeme_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lexeme_id UUID NOT NULL REFERENCES lexemes(id) ON DELETE CASCADE,
    form_id UUID REFERENCES lexeme_forms(id) ON DELETE SET NULL,
    media_asset_id UUID NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('image', 'audio_word', 'audio_sentence', 'audio_form', 'video')),
    label TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_lexeme_media_lexeme ON lexeme_media (lexeme_id);
