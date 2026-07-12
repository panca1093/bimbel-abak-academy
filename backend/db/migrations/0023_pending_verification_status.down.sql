UPDATE users SET status = 'active' WHERE status = 'pending_verification';

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_status_check;
ALTER TABLE users ADD CONSTRAINT users_status_check
    CHECK (status IN ('active', 'deactivated', 'deleted'));
