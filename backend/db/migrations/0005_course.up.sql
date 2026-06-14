CREATE TABLE IF NOT EXISTS course_section (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES product (id),
    title      TEXT NOT NULL,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS lesson (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id       UUID NOT NULL REFERENCES course_section (id),
    title            TEXT NOT NULL,
    video_url        TEXT,
    duration_seconds INT,
    position         INT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS course_enrollment (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users (id),
    product_id UUID NOT NULL REFERENCES product (id),
    order_id   UUID REFERENCES orders (id),
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    source     TEXT NOT NULL,
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_course_enrollment_student_product
    ON course_enrollment (student_id, product_id);

CREATE TABLE IF NOT EXISTS lesson_progress (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    enrollment_id UUID NOT NULL REFERENCES course_enrollment (id),
    lesson_id     UUID NOT NULL REFERENCES lesson (id),
    completed_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lesson_progress_enrollment_lesson
    ON lesson_progress (enrollment_id, lesson_id);
