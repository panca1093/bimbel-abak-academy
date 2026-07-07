-- FR-1: exam mode discriminator (standard | utbk | ielts). Existing rows read 'standard'.
ALTER TABLE exam
    ADD COLUMN mode TEXT NOT NULL DEFAULT 'standard'
    CHECK (mode IN ('standard', 'utbk', 'ielts'));

-- FR-2: Test section_type (nullable; IELTS L/R/W identity). Existing rows read NULL.
ALTER TABLE test
    ADD COLUMN section_type TEXT
    CHECK (section_type IS NULL OR section_type IN ('listening', 'reading', 'writing'));

-- FR-3: per-section session-timing state. One row per attached Test per session.
-- `submitted` is the terminal/locked state (no separate `locked` state).
CREATE TABLE exam_session_section (
    session_id       UUID        NOT NULL REFERENCES exam_session (id) ON DELETE CASCADE,
    test_id          UUID        NOT NULL REFERENCES test (id),
    sort_order       INT         NOT NULL,
    duration_minutes INT         NOT NULL,
    status           TEXT        NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'submitted')),
    started_at       TIMESTAMPTZ,
    submitted_at     TIMESTAMPTZ,
    extended_until   TIMESTAMPTZ,
    PRIMARY KEY (session_id, test_id)
);

CREATE INDEX idx_examsessionsection_order
    ON exam_session_section (session_id, sort_order);