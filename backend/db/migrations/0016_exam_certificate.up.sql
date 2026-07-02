-- Slice 5b+6 certificate fields: template choice on exam, generation timestamp on session.
ALTER TABLE exam ADD COLUMN IF NOT EXISTS certificate_template TEXT NOT NULL DEFAULT 'classic';

ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_certificate_template;
ALTER TABLE exam ADD CONSTRAINT chk_certificate_template
    CHECK (certificate_template IN ('classic', 'modern', 'elegant'));

ALTER TABLE exam_session ADD COLUMN IF NOT EXISTS certificate_generated_at TIMESTAMPTZ;
