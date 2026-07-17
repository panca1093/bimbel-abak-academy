-- Additive change; dropping loses data captured going forward.
ALTER TABLE users DROP COLUMN IF EXISTS unlisted_school_name;
