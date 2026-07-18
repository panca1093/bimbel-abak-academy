-- 0033_shipping_ongkir.up.sql
-- Add address hierarchy columns to orders table for Biteship Rates API quoting.
-- Depends on 0029_seed_province_city_district (province/city/district tables must exist).

ALTER TABLE orders ADD COLUMN province_id TEXT REFERENCES province(id);

ALTER TABLE orders ADD COLUMN city_id TEXT REFERENCES city(id);

ALTER TABLE orders ADD COLUMN district_id TEXT REFERENCES district(id);

ALTER TABLE orders ADD COLUMN kode_pos TEXT;
