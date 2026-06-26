UPDATE orders SET status = 'processing' WHERE status = 'shipped';

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('cart', 'payment_pending', 'paid', 'processing', 'completed', 'payment_expired', 'cancelled'));
