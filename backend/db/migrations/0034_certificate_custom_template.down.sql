UPDATE exam SET certificate_template = 'classic' WHERE certificate_template = 'custom';
ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_certificate_template;
ALTER TABLE exam ADD CONSTRAINT chk_certificate_template CHECK (certificate_template IN ('classic','modern','elegant'));
ALTER TABLE exam DROP COLUMN IF EXISTS certificate_background_url;
