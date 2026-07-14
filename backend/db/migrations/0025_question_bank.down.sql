-- Fail-safe down for 0025: reconstruct one test attachment per question,
-- never deleting question or exam_session_answer rows. Restores the FK's
-- ON DELETE CASCADE and uq_question_order to match the pre-0025 contract.
-- NOT NULL on test_id/sort_order is deliberately NOT restored: a bank-only
-- question (never attached to any test, impossible before 0025) has no
-- test_id to backfill, and this migration's own guarantee — proven by
-- TestMigration0025_QuestionBank — is that every question survives down,
-- including bank-only ones. Adding NOT NULL would make Postgres abort the
-- whole rollback the moment any bank-only question exists, which defeats
-- that guarantee. UNIQUE(test_id, sort_order) stays safe to restore because
-- Postgres treats NULLs as distinct, so multiple bank-only (NULL) rows don't
-- collide.

ALTER TABLE question ADD COLUMN IF NOT EXISTS test_id UUID REFERENCES test (id) ON DELETE CASCADE;
ALTER TABLE question ADD COLUMN IF NOT EXISTS sort_order INT;

-- Backfill each question from its first/only join row; when multiple
-- attachments exist, pick the lowest test_id to stay deterministic.
UPDATE question q
SET test_id = sq.test_id,
    sort_order = sq.sort_order
FROM (
    SELECT DISTINCT ON (question_id)
        question_id, test_id, sort_order
    FROM test_question
    ORDER BY question_id, test_id ASC
) sq
WHERE sq.question_id = q.id;

ALTER TABLE question ADD CONSTRAINT uq_question_order UNIQUE (test_id, sort_order);

DROP TABLE IF EXISTS test_question;

ALTER TABLE question DROP COLUMN IF EXISTS topic_id;

DROP TABLE IF EXISTS exam_topic;
