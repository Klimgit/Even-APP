CREATE TABLE IF NOT EXISTS demo_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO demo_notes (text) VALUES
    ('Привет из demo_notes'),
    ('Вторая заметка');
