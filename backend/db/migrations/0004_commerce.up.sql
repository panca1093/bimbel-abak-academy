CREATE TABLE IF NOT EXISTS product (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type            TEXT NOT NULL CHECK (type IN ('book', 'course', 'exam')),
    title           TEXT NOT NULL,
    description     TEXT,
    price           NUMERIC(12, 2) NOT NULL,
    stock           INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    is_visible      BOOLEAN NOT NULL DEFAULT false,
    weight_grams    INT,
    cover_image_url TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_product_type_status
    ON product (type, status);

CREATE TABLE IF NOT EXISTS promo_code (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                TEXT NOT NULL UNIQUE,
    discount_percent    NUMERIC(5, 2),
    discount_amount     NUMERIC(12, 2),
    min_order_amount    NUMERIC(12, 2),
    max_discount_amount NUMERIC(12, 2),
    max_uses            INT,
    uses                INT NOT NULL DEFAULT 0,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS orders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id          UUID NOT NULL REFERENCES users (id),
    status              TEXT NOT NULL DEFAULT 'cart' CHECK (status IN ('cart', 'payment_pending', 'paid', 'processing', 'shipped', 'cancelled', 'payment_expired', 'payment_failed')),
    subtotal            NUMERIC(12, 2),
    discount            NUMERIC(12, 2) DEFAULT 0,
    shipping_amount     NUMERIC(12, 2) DEFAULT 0,
    total               NUMERIC(12, 2),
    promo_code_id       UUID REFERENCES promo_code (id),
    shipping_address    JSONB,
    courier             TEXT,
    tracking_number     TEXT,
    shipped_at          TIMESTAMPTZ,
    payment_ref         TEXT,
    payment_expires_at  TIMESTAMPTZ,
    cancellation_reason TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_student_status
    ON orders (student_id, status);

CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders (status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_student_cart
    ON orders (student_id)
    WHERE status = 'cart';

CREATE TABLE IF NOT EXISTS order_item (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id     UUID NOT NULL REFERENCES orders (id),
    product_id   UUID NOT NULL REFERENCES product (id),
    product_type TEXT NOT NULL,
    title        TEXT NOT NULL,
    unit_price   NUMERIC(12, 2) NOT NULL,
    qty          INT NOT NULL DEFAULT 1,
    fulfilled_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_order_item_order
    ON order_item (order_id);
