-- Fail-safe down for 0026: reconstruct exam.product_id as a 1:1 link. Never deletes exam or
-- product rows. uq_exam_product enforces at most one exam per product_id, but the M:N window
-- this migration opened may have attached more than one exam to the same product — that can't
-- be expressed by a 1:1 column, so for each product_id we keep only the lowest exam_id and leave
-- the rest unlinked (product_id NULL, i.e. "free / not for sale" per the model's own contract).
-- This is the same lossy-but-safe collapse strategy 0025's down migration used for the
-- symmetric problem on the question/test_question side.

ALTER TABLE exam ADD COLUMN IF NOT EXISTS product_id UUID REFERENCES product (id);

UPDATE exam e
SET product_id = sub.product_id
FROM (
    SELECT DISTINCT ON (product_id)
        product_id, exam_id
    FROM product_exam
    ORDER BY product_id, exam_id ASC
) sub
WHERE sub.exam_id = e.id;

CREATE UNIQUE INDEX IF NOT EXISTS uq_exam_product
    ON exam (product_id)
    WHERE product_id IS NOT NULL;

DROP TABLE IF EXISTS product_exam;
