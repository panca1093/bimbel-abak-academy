-- Fail closed: an unverified account must never be promoted to active on rollback.
-- Lock it to deactivated so a rollback can't bypass the verification gate; an admin
-- can reactivate a legitimate one manually.
UPDATE users SET status = 'deactivated' WHERE status = 'pending_verification';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE users ADD CONSTRAINT users_status_check
    CHECK (status IN ('active', 'deactivated', 'deleted'));
