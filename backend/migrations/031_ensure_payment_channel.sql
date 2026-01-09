-- Ensure channel column exists in payment_orders table
-- This is a safety migration to ensure the column is added even if migration 030 had issues

-- Add channel column if it doesn't exist (idempotent)
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS channel VARCHAR(50) DEFAULT '';

-- Update existing records that don't have a channel set
UPDATE payment_orders
SET channel = CASE
    WHEN provider = 'zpay' THEN 'alipay'
    WHEN provider = 'stripe' THEN 'wechat'
    ELSE provider
END
WHERE channel IS NULL OR channel = '';

-- Add index if it doesn't exist (idempotent)
CREATE INDEX IF NOT EXISTS idx_payment_orders_channel ON payment_orders(channel);
