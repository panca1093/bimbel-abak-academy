-- Re-add 'shipped' to order status flow (physical fulfillment: processing → shipped → completed)
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('cart', 'payment_pending', 'paid', 'processing', 'shipped', 'completed', 'payment_expired', 'cancelled'));
