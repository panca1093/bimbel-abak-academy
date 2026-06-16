-- ─── product ───────────────────────────────────────────────────────────────
ALTER TABLE product RENAME COLUMN title TO name;
ALTER TABLE product RENAME COLUMN cover_image_url TO image_url;
ALTER TABLE product DROP COLUMN IF EXISTS is_visible;

ALTER TABLE product DROP CONSTRAINT IF EXISTS product_status_check;
ALTER TABLE product ADD CONSTRAINT product_status_check
    CHECK (status IN ('draft', 'published', 'hidden', 'archived'));

-- ─── orders ────────────────────────────────────────────────────────────────
ALTER TABLE orders RENAME COLUMN shipping_amount TO shipping_cost;
ALTER TABLE orders RENAME COLUMN payment_ref TO gateway_ref;
ALTER TABLE orders RENAME COLUMN courier TO selected_courier;

-- Drop and recreate status constraint to match TRD (remove shipped, payment_failed; add completed)
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_status_check
    CHECK (status IN ('cart', 'payment_pending', 'paid', 'processing', 'completed', 'payment_expired', 'cancelled'));

-- Migrate stale statuses before constraint takes effect
UPDATE orders SET status = 'processing' WHERE status = 'shipped';
UPDATE orders SET status = 'cancelled'  WHERE status = 'payment_failed';

ALTER TABLE orders ADD COLUMN IF NOT EXISTS paid_at                  TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS invoice_url              TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_method           TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS estimated_delivery_days  TEXT;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS checked_out_at           TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS completed_at             TIMESTAMPTZ;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancelled_at             TIMESTAMPTZ;

-- ─── order_item ────────────────────────────────────────────────────────────
ALTER TABLE order_item RENAME COLUMN title TO name;
ALTER TABLE order_item ADD COLUMN IF NOT EXISTS jumlah       NUMERIC(12,2);
ALTER TABLE order_item ADD COLUMN IF NOT EXISTS weight_grams INT;

-- Back-fill jumlah from existing rows
UPDATE order_item SET jumlah = unit_price * qty WHERE jumlah IS NULL;

-- ─── promo_code ────────────────────────────────────────────────────────────
ALTER TABLE promo_code RENAME COLUMN uses TO used_count;

-- ─── outbox ────────────────────────────────────────────────────────────────
ALTER TABLE outbox ADD COLUMN IF NOT EXISTS aggregate_type TEXT;

-- ─── webhook_log ────────────────────────────────────────────────────────────
ALTER TABLE webhook_log RENAME COLUMN payment_ref TO gateway_ref;
ALTER INDEX IF EXISTS idx_webhook_log_payment_ref RENAME TO idx_webhook_log_gateway_ref;

-- ─── job ───────────────────────────────────────────────────────────────────
-- Migrate status values before constraint change
UPDATE job SET status = 'queued'    WHERE status = 'pending';
UPDATE job SET status = 'succeeded' WHERE status = 'done';

ALTER TABLE job DROP CONSTRAINT IF EXISTS job_status_check;
ALTER TABLE job ADD CONSTRAINT job_status_check
    CHECK (status IN ('queued', 'running', 'succeeded', 'failed'));

ALTER TABLE job ADD COLUMN IF NOT EXISTS progress    INT  NOT NULL DEFAULT 0;
ALTER TABLE job ADD COLUMN IF NOT EXISTS result_url  TEXT;
ALTER TABLE job ADD COLUMN IF NOT EXISTS created_by  UUID REFERENCES users (id);

-- Drop the payload column (job input is no longer stored here; caller tracks context externally)
ALTER TABLE job DROP COLUMN IF EXISTS payload;
ALTER TABLE job DROP COLUMN IF EXISTS attempts;
ALTER TABLE job DROP COLUMN IF EXISTS last_error;
