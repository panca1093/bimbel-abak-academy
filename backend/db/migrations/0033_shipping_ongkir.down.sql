-- 0033_shipping_ongkir.down.sql
-- Reverse: drop address hierarchy columns from orders table.
-- WARNING: Dropped shipping address data cannot be restored -- this is irreversible data loss.

ALTER TABLE orders DROP COLUMN IF EXISTS kode_pos;

ALTER TABLE orders DROP COLUMN IF EXISTS district_id;

ALTER TABLE orders DROP COLUMN IF EXISTS city_id;

ALTER TABLE orders DROP COLUMN IF EXISTS province_id;
