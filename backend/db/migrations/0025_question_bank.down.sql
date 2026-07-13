-- Fail-safe down for 0025: reconstruct one test attachment per question,
-- never deleting question or exam_session_answer rows.

ALTER TABLE question ADD COLUMN IF NOT EXISTS test_id UUID REFERENCES test (id);
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

DROP TABLE IF EXISTS test_question;

ALTER TABLE question DROP COLUMN IF EXISTS topic_id;

DROP TABLE IF EXISTS exam_topic;
