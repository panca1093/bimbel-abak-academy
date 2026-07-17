-- 0030_user_biodata_changes.up.sql
-- Drop nis, add jenjang and address hierarchy reference fields.
-- Depends on 0029_seed_province_city_district (province/city/district tables must exist).

ALTER TABLE users DROP COLUMN nis;

ALTER TABLE users ADD COLUMN jenjang TEXT;

ALTER TABLE users ADD COLUMN provinsi_id TEXT REFERENCES province(id);

ALTER TABLE users ADD COLUMN kota_id TEXT REFERENCES city(id);

ALTER TABLE users ADD COLUMN kecamatan_id TEXT REFERENCES district(id);

ALTER TABLE users ADD COLUMN kode_pos TEXT;
