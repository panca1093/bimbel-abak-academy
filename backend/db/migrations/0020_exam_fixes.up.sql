-- timer_mode vocabulary drift: schema.dbml/TRD/PRD (FR-EXAM-10) say 'per_test';
-- slices 2-7 shipped 'per_question'. Rename stored values to match the docs.
UPDATE exam SET timer_mode = 'per_test' WHERE timer_mode = 'per_question';

-- Exams created before the service defaulted status were stored with '' instead
-- of 'draft' (the INSERT wrote the zero value, bypassing the column DEFAULT).
UPDATE exam SET status = 'draft' WHERE status = '';
