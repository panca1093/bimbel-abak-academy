-- Drop the order_participant table. Participant-fan-out metadata is lost on rollback.
DROP TABLE IF EXISTS order_participant;
