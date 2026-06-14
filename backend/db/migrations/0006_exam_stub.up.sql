CREATE TABLE IF NOT EXISTS exam_registration (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID NOT NULL REFERENCES users (id),
    exam_id     UUID NOT NULL,
    order_id    UUID REFERENCES orders (id),
    token       TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    card_pdf_url TEXT,
    checked_in_at TIMESTAMPTZ,
    status      TEXT NOT NULL DEFAULT 'registered' CHECK (status IN ('registered', 'expired')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_registration_student_exam_order
    ON exam_registration (student_id, exam_id, order_id);
