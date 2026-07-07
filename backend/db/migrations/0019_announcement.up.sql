CREATE TABLE IF NOT EXISTS announcement (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL,
    message         TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('announcement', 'promo', 'exam')),
    recipients      TEXT NOT NULL CHECK (recipients IN ('all', 'students', 'admins')),
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sent')),
    scheduled_at    TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    recipient_count INT,
    created_by      UUID REFERENCES users (id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_announcement_status_scheduled
    ON announcement (status, scheduled_at);
