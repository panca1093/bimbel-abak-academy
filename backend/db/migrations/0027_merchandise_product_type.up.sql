-- Widen product.type to allow physical merchandise alongside book/course/exam.
-- The CHECK was defined inline in 0004_commerce.up.sql; Postgres auto-named it
-- product_type_check. TEXT+CHECK per repo convention, no native enum.
ALTER TABLE product DROP CONSTRAINT IF EXISTS product_type_check;
ALTER TABLE product ADD CONSTRAINT product_type_check
    CHECK (type IN ('book', 'course', 'exam', 'merchandise'));
