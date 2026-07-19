-- 0034_shipping_selected_service.up.sql
-- Persist the specific courier service selected for a shipment (e.g. "REG"
-- vs "YES" within JNE), not just the carrier name, so PatchCart can price
-- against the exact quoted rate instead of the first same-carrier match.

ALTER TABLE orders ADD COLUMN selected_service TEXT;
