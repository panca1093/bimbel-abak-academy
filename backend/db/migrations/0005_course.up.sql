CREATE TABLE IF NOT EXISTS course (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL,
    level           TEXT,
    subject         TEXT,
    instructor_name TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_course (
    product_id UUID NOT NULL REFERENCES product (id),
    course_id  UUID NOT NULL REFERENCES course (id),
    PRIMARY KEY (product_id, course_id)
);

CREATE TABLE IF NOT EXISTS section (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id  UUID NOT NULL REFERENCES course (id),
    title      TEXT NOT NULL,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lesson (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id       UUID NOT NULL REFERENCES section (id),
    title            TEXT NOT NULL,
    video_url        TEXT,
    duration_seconds INT,
    position         INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS course_session (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES users (id),
    course_id         UUID NOT NULL REFERENCES course (id),
    order_id          UUID REFERENCES orders (id),
    status            TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    source            TEXT NOT NULL,
    enrolled_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at        TIMESTAMPTZ,
    completed_lessons JSONB NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_course_session_student_course
    ON course_session (student_id, course_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_section_course_position
    ON section (course_id, position);

CREATE INDEX IF NOT EXISTS idx_lesson_section_position
    ON lesson (section_id, position);

CREATE INDEX IF NOT EXISTS idx_course_session_order
    ON course_session (order_id);

CREATE INDEX IF NOT EXISTS idx_product_course_course
    ON product_course (course_id);
