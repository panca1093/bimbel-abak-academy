-- Migration 0025: reusable question bank — topics, test_question join, question repoint.
-- FR-1..FR-8 of the Exam Admin mockup-parity + reusable question bank spec.

CREATE TABLE IF NOT EXISTS exam_topic (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    subject    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_exam_topic_subject_name UNIQUE (subject, name)
);

CREATE TABLE IF NOT EXISTS test_question (
    test_id     UUID NOT NULL REFERENCES test (id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    sort_order  INT  NOT NULL,
    PRIMARY KEY (test_id, question_id),
    CONSTRAINT uq_test_question_order UNIQUE (test_id, sort_order)
);

-- Distinct (subject, topic) from existing tests become curated topics.
INSERT INTO exam_topic (name, subject)
SELECT DISTINCT topic, subject
FROM test;

-- Add the new topic reference before populating it.
ALTER TABLE question ADD COLUMN IF NOT EXISTS topic_id UUID REFERENCES exam_topic (id);

-- One join row per existing question, preserving its current test and order.
INSERT INTO test_question (test_id, question_id, sort_order)
SELECT test_id, id, sort_order
FROM question;

-- Point each question at the topic derived from its current test.
UPDATE question q
SET topic_id = (
    SELECT t.id
    FROM exam_topic t
    JOIN test tt ON tt.subject = t.subject AND tt.topic = t.name
    WHERE tt.id = q.test_id
);

-- Drop the old per-question test coupling.
ALTER TABLE question DROP COLUMN IF EXISTS test_id;
ALTER TABLE question DROP COLUMN IF EXISTS sort_order;
ALTER TABLE question DROP CONSTRAINT IF EXISTS uq_question_order;
