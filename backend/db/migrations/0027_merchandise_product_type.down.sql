-- Fail-safe: convert any merchandise rows to 'book' (closest physical type)
-- BEFORE re-adding the narrow CHECK, so down is always runnable and never
-- orphans rows.
UPDATE product SET type = 'book' WHERE type = 'merchandise';

ALTER TABLE product DROP CONSTRAINT IF EXISTS product_type_check;
ALTER TABLE product ADD CONSTRAINT product_type_check
    CHECK (type IN ('book', 'course', 'exam'));
