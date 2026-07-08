-- Reverts 0021: drop section table, then the two column adds.
DROP TABLE IF EXISTS exam_session_section;

ALTER TABLE test
    DROP COLUMN IF EXISTS section_type;

ALTER TABLE exam
    DROP COLUMN IF EXISTS mode;