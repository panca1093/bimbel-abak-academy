-- ─── job ───────────────────────────────────────────────────────────────────
ALTER TABLE job DROP COLUMN IF EXISTS progress;
ALTER TABLE job DROP COLUMN IF EXISTS result_url;
ALTER TABLE job DROP COLUMN IF EXISTS created_by;
ALTER TABLE job ADD COLUMN IF NOT EXISTS payload    JSONB NOT NULL DEFAULT '{}';
ALTER TABLE job ADD COLUMN IF NOT EXISTS attempts   INT NOT NULL DEFAULT 0;
ALTER TABLE job ADD COLUMN IF NOT EXISTS last_error TEXT;

ALTER TABLE job DROP CONSTRAINT IF EXISTS job_status_check;
ALTER TABLE job ADD CONSTRAINT job_status_check
    CHECK (status IN ('pending', 'running', 'done', 'failed'));
UPDATE job SET status = 'pending' WHERE status = 'queued';
UPDATE job SET status = 'done'    WHERE status = 'succeeded';

-- ─── webhook_log ────────────────────────────────────────────────────────────
ALTER INDEX IF EXISTS idx_webhook_log_gateway_ref RENAME TO idx_webhook_log_payment_ref;
ALTER TABLE webhook_log RENAME COLUMN gateway_ref TO payment_ref;

-- ─── outbox ────────────────────────────────────────────────────────────────
ALTER TABLE outbox DROP COLUMN IF EXISTS aggregate_type;

-- ─── promo_code ────────────────────────────────────────────────────────────
ALTER TABLE promo_code RENAME COLUMN used_count TO uses;

-- ─── order_item ────────────────────────────────────────────────────────────
ALTER TABLE order_item DROP COLUMN IF EXISTS jumlah;
ALTER TABLE order_item DROP COLUMN IF EXISTS weight_grams;
ALTER TABLE order_item RENAME COLUMN name TO title;

-- ─── orders ────────────────────────────────────────────────────────────────
ALTER TABLE orders DROP COLUMN IF EXISTS paid_at;
ALTER TABLE orders DROP COLUMN IF EXISTS invoice_url;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_method;
ALTER TABLE orders DROP COLUMN IF EXISTS estimated_delivery_days;
ALTER TABLE orders DROP COLUMN IF EXISTS checked_out_at;
ALTER TABLE orders DROP COLUMN IF EXISTS completed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS cancelled_at;

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('cart', 'payment_pending', 'paid', 'processing', 'shipped', 'cancelled', 'payment_expired', 'payment_failed'));

ALTER TABLE orders RENAME COLUMN selected_courier TO courier;
ALTER TABLE orders RENAME COLUMN gateway_ref TO payment_ref;
ALTER TABLE orders RENAME COLUMN shipping_cost TO shipping_amount;

-- ─── product ───────────────────────────────────────────────────────────────
ALTER TABLE product DROP CONSTRAINT IF EXISTS product_status_check;
ALTER TABLE product ADD CONSTRAINT product_status_check
    CHECK (status IN ('draft', 'published', 'archived'));

ALTER TABLE product ADD COLUMN IF NOT EXISTS is_visible BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE product RENAME COLUMN image_url TO cover_image_url;
ALTER TABLE product RENAME COLUMN name TO title;
