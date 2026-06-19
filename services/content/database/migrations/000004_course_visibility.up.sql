ALTER TABLE courses
    ADD COLUMN visibility TEXT NOT NULL DEFAULT 'invite_only'
        CHECK (visibility IN ('public', 'invite_only'));
