-- Keep v0.1.126 payment order reads/writes compatible with databases that
-- already contain the legacy dev payment_orders table.
--
-- This migration is intentionally additive:
--   * legacy dev columns are added to fresh v0.1.126 installs as nullable/defaulted
--     columns so old data can coexist with new code;
--   * v0.1.126 Ent columns are added/backfilled on legacy dev databases because
--     092_payment_orders.sql is CREATE TABLE IF NOT EXISTS and does not alter an
--     existing legacy table.

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS user_email VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS user_name VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS user_notes TEXT,
    ADD COLUMN IF NOT EXISTS amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pay_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee_rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS recharge_code VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS out_trade_no VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_type VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pay_url TEXT,
    ADD COLUMN IF NOT EXISTS qr_code TEXT,
    ADD COLUMN IF NOT EXISTS qr_code_img TEXT,
    ADD COLUMN IF NOT EXISTS order_type VARCHAR(20) NOT NULL DEFAULT 'balance',
    ADD COLUMN IF NOT EXISTS plan_id BIGINT,
    ADD COLUMN IF NOT EXISTS subscription_group_id BIGINT,
    ADD COLUMN IF NOT EXISTS subscription_days INT,
    ADD COLUMN IF NOT EXISTS provider_instance_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS provider_key VARCHAR(30),
    ADD COLUMN IF NOT EXISTS provider_snapshot JSONB,
    ADD COLUMN IF NOT EXISTS refund_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS refund_reason TEXT,
    ADD COLUMN IF NOT EXISTS refund_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS force_refund BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS refund_requested_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS refund_request_reason TEXT,
    ADD COLUMN IF NOT EXISTS refund_requested_by VARCHAR(20),
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed_reason TEXT,
    ADD COLUMN IF NOT EXISTS client_ip VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS src_host VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS src_url TEXT,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS order_no VARCHAR(64) DEFAULT '',
    ADD COLUMN IF NOT EXISTS trade_no VARCHAR(128),
    ADD COLUMN IF NOT EXISTS username VARCHAR(100),
    ADD COLUMN IF NOT EXISTS amount_cny DECIMAL(20,2) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS amount_usd DECIMAL(20,8) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bonus_usd DECIMAL(20,8) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_usd DECIMAL(20,8) DEFAULT 0,
    ADD COLUMN IF NOT EXISTS exchange_rate DECIMAL(10,4) DEFAULT 1,
    ADD COLUMN IF NOT EXISTS provider VARCHAR(30) DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50),
    ADD COLUMN IF NOT EXISTS payment_url TEXT,
    ADD COLUMN IF NOT EXISTS expire_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS remark VARCHAR(255) DEFAULT '',
    ADD COLUMN IF NOT EXISTS channel VARCHAR(50) DEFAULT '',
    ADD COLUMN IF NOT EXISTS biz_type VARCHAR(40) DEFAULT 'online_recharge',
    ADD COLUMN IF NOT EXISTS biz_group_id BIGINT,
    ADD COLUMN IF NOT EXISTS biz_validity_days INT,
    ADD COLUMN IF NOT EXISTS promotion_tier INT,
    ADD COLUMN IF NOT EXISTS promotion_used BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS callback_data TEXT,
    ADD COLUMN IF NOT EXISTS callback_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS user_agent VARCHAR(500);

DO $$
DECLARE
    spec record;
BEGIN
    FOR spec IN
        SELECT * FROM (VALUES
            ('order_no',      ''''''),
            ('amount_cny',    '0'),
            ('amount_usd',    '0'),
            ('bonus_usd',     '0'),
            ('total_usd',     '0'),
            ('exchange_rate', '1'),
            ('provider',      ''''''),
            ('remark',        ''''''),
            ('channel',       ''''''),
            ('biz_type',      '''online_recharge'''),
            ('promotion_used','FALSE')
        ) AS t(col, default_expr)
    LOOP
        EXECUTE format('ALTER TABLE payment_orders ALTER COLUMN %I SET DEFAULT %s', spec.col, spec.default_expr);
        EXECUTE format('UPDATE payment_orders SET %I = DEFAULT WHERE %I IS NULL', spec.col, spec.col);
    END LOOP;
END
$$;

DO $$
DECLARE
    legacy_col text;
BEGIN
    FOREACH legacy_col IN ARRAY ARRAY[
        'order_no',
        'amount_cny',
        'amount_usd',
        'bonus_usd',
        'total_usd',
        'exchange_rate',
        'provider',
        'expire_at'
    ]
    LOOP
        IF EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = current_schema()
              AND table_name = 'payment_orders'
              AND column_name = legacy_col
              AND is_nullable = 'NO'
        ) THEN
            EXECUTE format('ALTER TABLE payment_orders ALTER COLUMN %I DROP NOT NULL', legacy_col);
        END IF;
    END LOOP;
END
$$;

ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_order_no_key;
DROP INDEX IF EXISTS payment_orders_order_no_key;
CREATE UNIQUE INDEX IF NOT EXISTS payment_orders_order_no_uniq
    ON payment_orders (order_no)
    WHERE order_no IS NOT NULL AND order_no <> '';

UPDATE payment_orders
SET client_ip = ''
WHERE client_ip IS NULL;

UPDATE payment_orders AS po
SET user_email = COALESCE(NULLIF(BTRIM(po.user_email), ''), u.email, ''),
    user_name = COALESCE(NULLIF(BTRIM(po.user_name), ''), NULLIF(BTRIM(po.username), ''), u.username, ''),
    user_notes = COALESCE(NULLIF(po.user_notes, ''), NULLIF(u.notes, ''))
FROM users AS u
WHERE u.id = po.user_id;

UPDATE payment_orders
SET out_trade_no = CASE
        WHEN BTRIM(COALESCE(out_trade_no, '')) = '' THEN COALESCE(NULLIF(BTRIM(order_no), ''), out_trade_no, '')
        ELSE out_trade_no
    END,
    amount = CASE
        WHEN COALESCE(amount, 0) = 0 AND COALESCE(amount_cny, 0) > 0 THEN amount_cny
        ELSE amount
    END,
    pay_amount = CASE
        WHEN COALESCE(pay_amount, 0) = 0 AND COALESCE(amount_cny, 0) > 0 THEN amount_cny
        ELSE pay_amount
    END,
    recharge_code = CASE
        WHEN BTRIM(COALESCE(recharge_code, '')) = '' THEN COALESCE(NULLIF(BTRIM(order_no), ''), recharge_code, '')
        ELSE recharge_code
    END,
    payment_type = CASE
        WHEN BTRIM(COALESCE(payment_type, '')) = '' THEN COALESCE(NULLIF(BTRIM(channel), ''), NULLIF(BTRIM(provider), ''), NULLIF(BTRIM(payment_method), ''), '')
        ELSE payment_type
    END,
    payment_trade_no = CASE
        WHEN BTRIM(COALESCE(payment_trade_no, '')) = '' THEN COALESCE(NULLIF(BTRIM(trade_no), ''), '')
        ELSE payment_trade_no
    END,
    pay_url = COALESCE(pay_url, NULLIF(payment_url, '')),
    order_type = CASE
        WHEN BTRIM(COALESCE(order_type, '')) = '' OR order_type = 'balance' THEN
            CASE WHEN COALESCE(biz_type, 'online_recharge') = 'subscription_purchase' THEN 'subscription' ELSE 'balance' END
        ELSE order_type
    END,
    subscription_group_id = COALESCE(subscription_group_id, biz_group_id),
    subscription_days = COALESCE(subscription_days, biz_validity_days),
    expires_at = COALESCE(expire_at, expires_at),
    completed_at = COALESCE(
        completed_at,
        CASE
            WHEN LOWER(COALESCE(status, '')) IN ('paid', 'completed', 'refunded') THEN paid_at
            ELSE NULL
        END
    ),
    updated_at = COALESCE(updated_at, NOW());

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_created
    ON payment_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_status_created
    ON payment_orders(user_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status_created
    ON payment_orders(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_channel
    ON payment_orders(channel);
CREATE INDEX IF NOT EXISTS idx_payment_orders_biz_type_created
    ON payment_orders(biz_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_orders_biz_group_id
    ON payment_orders(biz_group_id)
    WHERE biz_group_id IS NOT NULL;
