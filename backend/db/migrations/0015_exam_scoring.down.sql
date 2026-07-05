ALTER TABLE exam DROP CONSTRAINT IF EXISTS chk_result_config;

ALTER TABLE question DROP COLUMN IF EXISTS point_wrong;
ALTER TABLE question DROP COLUMN IF EXISTS point_correct;
