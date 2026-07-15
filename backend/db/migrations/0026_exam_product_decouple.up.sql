-- Migration 0026: decouple Exam from Product (M:N via product_exam), mirroring product_course.
-- TRD line 959's "Same pattern applies to Exam.product_id" was never implemented until now;
-- CreateProductAndExamTx's auto-create-1:1-on-exam-create is retired in the same change —
-- Exam creation no longer implicitly creates a Product, and a Product can now attach more
-- than one Exam (same as course-type products already can via product_course).

CREATE TABLE IF NOT EXISTS product_exam (
    product_id UUID NOT NULL REFERENCES product (id),
    exam_id    UUID NOT NULL REFERENCES exam (id),
    PRIMARY KEY (product_id, exam_id)
);

CREATE INDEX IF NOT EXISTS idx_product_exam_exam
    ON product_exam (exam_id);

-- One join row per exam that already has a linked product, preserving the existing 1:1 link.
INSERT INTO product_exam (product_id, exam_id)
SELECT product_id, id
FROM exam
WHERE product_id IS NOT NULL;

DROP INDEX IF EXISTS uq_exam_product;
ALTER TABLE exam DROP COLUMN IF EXISTS product_id;
