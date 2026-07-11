-- audit_log.actor_id must always name the actor. The only historical NULL
-- producer was AdminRefundOrder (now fixed to attribute a real actor); those
-- rows record a refund without who performed it, so they carry no
-- accountability value and are removed before the constraint is added.
-- System-originated audit (no human actor) does not exist yet — when it does,
-- add actor_type ('user' | 'system') and a CHECK rather than a sentinel UUID.
DELETE FROM audit_log WHERE actor_id IS NULL;

ALTER TABLE audit_log ALTER COLUMN actor_id SET NOT NULL;
