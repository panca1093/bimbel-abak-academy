ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_certificate_template;
ALTER TABLE exam ADD CONSTRAINT chk_certificate_template CHECK (certificate_template IN ('classic','modern','elegant','custom'));
ALTER TABLE exam ADD COLUMN IF NOT EXISTS certificate_background_url TEXT;
