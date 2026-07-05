ALTER TABLE job ADD COLUMN IF NOT EXISTS input_url TEXT;
ALTER TABLE job ADD COLUMN IF NOT EXISTS error      TEXT;

-- 0009 changed job_status_check to ('queued', 'running', 'succeeded', 'failed')
-- but left this column's DEFAULT at 'pending' (from 0007), so any insert that
-- relies on the default (e.g. CreateJob) violates the constraint.
ALTER TABLE job ALTER COLUMN status SET DEFAULT 'queued';
