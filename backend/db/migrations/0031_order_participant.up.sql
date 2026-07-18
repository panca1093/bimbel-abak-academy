-- Create order_participant table to support bulk exam orders.
-- Each row represents one student participant in a bulk exam purchase.
-- Presence of rows indicates fan-out (register each participant) vs.
-- zero rows = today's behavior (register the buyer, order.student_id).
CREATE TABLE IF NOT EXISTS order_participant (
    order_id   UUID NOT NULL REFERENCES orders (id),
    student_id UUID NOT NULL REFERENCES users (id),
    PRIMARY KEY (order_id, student_id)
);
