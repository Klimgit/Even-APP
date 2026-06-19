CREATE TABLE IF NOT EXISTS audit_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    UUID,
    action      TEXT NOT NULL,
    target_type TEXT,
    target_id   UUID,
    details     JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_created ON audit_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit_events (action);
