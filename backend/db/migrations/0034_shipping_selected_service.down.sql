-- 0034_shipping_selected_service.down.sql
-- Additive change; dropping loses data captured going forward.
ALTER TABLE orders DROP COLUMN IF EXISTS selected_service;
