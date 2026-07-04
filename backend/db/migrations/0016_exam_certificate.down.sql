ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_certificate_template;

ALTER TABLE exam DROP COLUMN IF EXISTS certificate_template;

ALTER TABLE exam_session DROP COLUMN IF EXISTS certificate_generated_at;
