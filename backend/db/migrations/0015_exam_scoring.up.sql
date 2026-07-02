-- Slice 5 scoring: per-question points (author-facing magnitudes; the scoring engine
-- applies the sign) and a CHECK on exam.result_config values.
ALTER TABLE question ADD COLUMN IF NOT EXISTS point_correct INT NOT NULL DEFAULT 1 CHECK (point_correct >= 1);
ALTER TABLE question ADD COLUMN IF NOT EXISTS point_wrong   INT NOT NULL DEFAULT 0 CHECK (point_wrong >= 0);

ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_result_config;
ALTER TABLE exam ADD CONSTRAINT chk_result_config
    CHECK (result_config IN ('hidden', 'score_only', 'score_pembahasan'));
