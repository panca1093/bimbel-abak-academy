CREATE UNIQUE INDEX idx_orders_gateway_ref ON orders(gateway_ref) WHERE gateway_ref IS NOT NULL;
