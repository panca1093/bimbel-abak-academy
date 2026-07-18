-- 0030_user_biodata_changes.down.sql
-- Reverse: drop address fields, restore nis.
-- WARNING: Dropped nis values cannot be restored -- this is irreversible data loss.
-- Matches NFR-01: same class of accepted data loss as 0027's merchandise/medal collapse.

ALTER TABLE users DROP COLUMN IF EXISTS kode_pos;
ALTER TABLE users DROP COLUMN IF EXISTS kecamatan_id;
ALTER TABLE users DROP COLUMN IF EXISTS kota_id;
ALTER TABLE users DROP COLUMN IF EXISTS provinsi_id;
ALTER TABLE users DROP COLUMN IF EXISTS jenjang;

ALTER TABLE users ADD COLUMN nis TEXT;
