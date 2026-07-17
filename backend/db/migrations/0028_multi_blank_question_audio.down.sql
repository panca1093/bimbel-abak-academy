-- Fail-safe: convert multi_blank rows to fill_blank before re-narrowing the CHECK,
-- so down is always runnable and never orphans rows. Per-blank answers in question_blank
-- do not survive this rollback (documented, accepted data loss per FR-2).

UPDATE question SET format = 'fill_blank' WHERE format = 'multi_blank';

-- Drop the question_blank table. Per-blank correct_answer data is lost on rollback.
DROP TABLE IF EXISTS question_blank;

-- Drop the audio_url column added in the up migration.
ALTER TABLE question DROP COLUMN IF EXISTS audio_url;

-- Restore the original narrower CHECK to only accept the 5 pre-multi_blank formats.
ALTER TABLE question DROP CONSTRAINT IF EXISTS question_format_check;
ALTER TABLE question ADD CONSTRAINT question_format_check
    CHECK (format IN ('mcq', 'multi_answer', 'short', 'fill_blank', 'essay'));
