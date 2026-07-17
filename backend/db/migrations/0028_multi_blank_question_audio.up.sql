-- Add multi_blank format to question.format CHECK and create question_blank table
-- for per-blank answers. Also add optional audio_url for per-question listening clips.

-- Widen question.format CHECK to allow 'multi_blank' alongside the existing 5 formats.
ALTER TABLE question DROP CONSTRAINT IF EXISTS question_format_check;
ALTER TABLE question ADD CONSTRAINT question_format_check
    CHECK (format IN ('mcq', 'multi_answer', 'short', 'fill_blank', 'essay', 'multi_blank'));

-- Create question_blank table to store per-blank correct answers.
-- Mirrors question_option's composite-PK shape, no surrogate ID.
CREATE TABLE IF NOT EXISTS question_blank (
    question_id    UUID NOT NULL REFERENCES question (id) ON DELETE CASCADE,
    blank_index    INT  NOT NULL,
    correct_answer TEXT NOT NULL,
    PRIMARY KEY (question_id, blank_index)
);

-- Add optional audio_url to question for per-question listening clips.
ALTER TABLE question ADD COLUMN IF NOT EXISTS audio_url TEXT;
