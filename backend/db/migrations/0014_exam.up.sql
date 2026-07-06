-- Exam module — all 9 tables from requirements/schema.dbml lines 233–386
-- Slice 0 lays down schema for Test + Question + QuestionOption (authoring);
-- the rest are created now so later slices are pure logic.

CREATE TABLE IF NOT EXISTS test (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title            TEXT NOT NULL,
    subject          TEXT NOT NULL,
    topic            TEXT NOT NULL,
    duration_minutes INT  NOT NULL,
    audio_url        TEXT,
    audio_play_limit INT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS question (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id        UUID NOT NULL REFERENCES test (id) ON DELETE CASCADE,
    format         TEXT NOT NULL CHECK (format IN ('mcq', 'multi_answer', 'short', 'fill_blank', 'essay')),
    body           TEXT NOT NULL,
    correct_answer TEXT,
    explanation    TEXT,
    difficulty     TEXT,
    image_url      TEXT,
    sort_order     INT  NOT NULL,
    CONSTRAINT uq_question_order UNIQUE (test_id, sort_order)
);

CREATE TABLE IF NOT EXISTS question_option (
    question_id UUID    NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    key         TEXT    NOT NULL,
    text        TEXT    NOT NULL,
    image_url   TEXT,
    is_correct  BOOLEAN NOT NULL DEFAULT false,
    sort_order  INT     NOT NULL,
    PRIMARY KEY (question_id, key),
    CONSTRAINT uq_questionoption_order UNIQUE (question_id, sort_order)
);

CREATE TABLE IF NOT EXISTS exam (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title                   TEXT        NOT NULL,
    is_free                 BOOLEAN     NOT NULL DEFAULT false,
    scheduled_at            TIMESTAMPTZ,
    requires_checkin        BOOLEAN     NOT NULL DEFAULT false,
    allow_leaderboard       BOOLEAN     NOT NULL DEFAULT false,
    cdn_bundle              BOOLEAN     NOT NULL DEFAULT false,
    bundle_url              TEXT,
    bundle_generated_at     TIMESTAMPTZ,
    check_in_window_minutes INT,
    grace_window_minutes    INT,
    max_attempts            INT,
    timer_mode              TEXT        NOT NULL DEFAULT 'overall',
    duration_minutes        INT,
    randomize               BOOLEAN     NOT NULL DEFAULT false,
    result_config           TEXT        NOT NULL DEFAULT 'hidden',
    result_release_at       TIMESTAMPTZ,
    status                  TEXT        NOT NULL DEFAULT 'draft',
    product_id              UUID        REFERENCES product (id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exam_status
    ON exam (status);

CREATE INDEX IF NOT EXISTS idx_exam_schedule
    ON exam (scheduled_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_exam_product
    ON exam (product_id)
    WHERE product_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS exam_test (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id    UUID NOT NULL REFERENCES exam (id),
    test_id    UUID NOT NULL REFERENCES test (id),
    sort_order INT  NOT NULL,
    CONSTRAINT uq_examtest UNIQUE (exam_id, test_id)
);

CREATE INDEX IF NOT EXISTS idx_examtest_order
    ON exam_test (exam_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_examtest_test
    ON exam_test (test_id);

CREATE TABLE IF NOT EXISTS exam_registration (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id     UUID        NOT NULL REFERENCES users (id),
    exam_id        UUID        NOT NULL REFERENCES exam (id),
    token          TEXT        NOT NULL,
    card_pdf_url   TEXT,
    checked_in_at  TIMESTAMPTZ,
    attempts_used  INT         NOT NULL DEFAULT 0,
    status         TEXT        NOT NULL DEFAULT 'registered',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_examregistration UNIQUE (student_id, exam_id),
    CONSTRAINT uq_examregistration_token UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS idx_examregistration_status
    ON exam_registration (exam_id, status);

CREATE TABLE IF NOT EXISTS exam_session (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID         NOT NULL REFERENCES exam_registration (id),
    student_id      UUID         NOT NULL REFERENCES users (id),
    exam_id         UUID         NOT NULL REFERENCES exam (id),
    attempt_number  INT          NOT NULL DEFAULT 1,
    started_at      TIMESTAMPTZ  NOT NULL,
    submitted_at    TIMESTAMPTZ,
    extended_until  TIMESTAMPTZ,
    admin_submitted BOOLEAN      NOT NULL DEFAULT false,
    score           NUMERIC(6, 2),
    certificate_url TEXT,
    last_saved_at   TIMESTAMPTZ,
    status          TEXT         NOT NULL DEFAULT 'in_progress',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_examsession_registration
    ON exam_session (registration_id);

CREATE INDEX IF NOT EXISTS idx_examsession_monitor
    ON exam_session (exam_id, status);

CREATE INDEX IF NOT EXISTS idx_examsession_student
    ON exam_session (student_id, exam_id);

CREATE TABLE IF NOT EXISTS exam_session_answer (
    session_id         UUID         NOT NULL REFERENCES exam_session (id) ON DELETE CASCADE,
    question_id        UUID         NOT NULL REFERENCES question (id),
    answer             TEXT,
    is_correct         BOOLEAN,
    score              NUMERIC(6, 2),
    graded_by          UUID         REFERENCES users (id),
    graded_at          TIMESTAMPTZ,
    grader_comment     TEXT,
    flagged_for_review BOOLEAN      NOT NULL DEFAULT false,
    saved_at           TIMESTAMPTZ  NOT NULL,
    PRIMARY KEY (session_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_examsessionanswer_grading
    ON exam_session_answer (session_id)
    WHERE graded_at IS NULL;

CREATE TABLE IF NOT EXISTS session_violation_log (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id     UUID        NOT NULL REFERENCES exam_session (id),
    student_id     UUID        NOT NULL REFERENCES users (id),
    violation_type TEXT        NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_violationlog_session
    ON session_violation_log (session_id);
