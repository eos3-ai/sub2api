-- +goose Up
-- +goose StatementBegin
-- Add channel column to payment_orders table
-- This allows storing the original payment channel (alipay/wechat) separately from the provider (zpay/stripe)

ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS channel VARCHAR(50) DEFAULT '';

-- Update existing records: for zpay provider, set channel to 'alipay' (default assumption)
UPDATE payment_orders SET channel = 'alipay' WHERE provider = 'zpay' AND (channel IS NULL OR channel = '');

-- For stripe provider, set channel to 'wechat' (current usage)
UPDATE payment_orders SET channel = 'wechat' WHERE provider = 'stripe' AND (channel IS NULL OR channel = '');

-- For other providers, set channel same as provider
UPDATE payment_orders SET channel = provider WHERE channel IS NULL OR channel = '';

-- Add index for channel queries
CREATE INDEX IF NOT EXISTS idx_payment_orders_channel ON payment_orders(channel);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove channel column and index
DROP INDEX IF EXISTS idx_payment_orders_channel;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS channel;
-- +goose StatementEnd
