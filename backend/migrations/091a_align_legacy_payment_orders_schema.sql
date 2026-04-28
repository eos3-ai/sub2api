-- Align legacy payment_orders rows/table with the v0.1.119 schema baseline.
-- This must run before 092_payment_orders.sql because older databases already
-- have payment_orders from 008_payment_order.sql but are missing the newer
-- columns/index targets introduced by the Ent-based payment subsystem.

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS user_email VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS user_name VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS user_notes TEXT,
    ADD COLUMN IF NOT EXISTS amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pay_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee_rate DECIMAL(10,4) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS recharge_code VARCHAR(64) NOT NULL DEFAULT '',
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
    ADD COLUMN IF NOT EXISTS src_host VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS src_url TEXT;

UPDATE payment_orders
SET client_ip = ''
WHERE client_ip IS NULL;

UPDATE payment_orders AS po
SET user_email = COALESCE(u.email, po.user_email, ''),
    user_name = CASE
        WHEN BTRIM(COALESCE(po.user_name, '')) <> '' THEN po.user_name
        WHEN BTRIM(COALESCE(po.username, '')) <> '' THEN po.username
        ELSE COALESCE(u.username, '')
    END,
    user_notes = CASE
        WHEN po.user_notes IS NOT NULL AND BTRIM(po.user_notes) <> '' THEN po.user_notes
        ELSE NULLIF(u.notes, '')
    END
FROM users AS u
WHERE u.id = po.user_id;

UPDATE payment_orders
SET user_name = COALESCE(NULLIF(BTRIM(user_name), ''), COALESCE(username, ''), ''),
    amount = CASE WHEN amount = 0 THEN COALESCE(amount_cny, 0) ELSE amount END,
    pay_amount = CASE WHEN pay_amount = 0 THEN COALESCE(amount_cny, 0) ELSE pay_amount END,
    recharge_code = CASE
        WHEN BTRIM(COALESCE(recharge_code, '')) = '' THEN COALESCE(order_no, '')
        ELSE recharge_code
    END,
    payment_type = CASE
        WHEN BTRIM(COALESCE(payment_type, '')) = '' THEN COALESCE(NULLIF(channel, ''), NULLIF(provider, ''), '')
        ELSE payment_type
    END,
    payment_trade_no = CASE
        WHEN BTRIM(COALESCE(payment_trade_no, '')) = '' THEN COALESCE(trade_no, '')
        ELSE payment_trade_no
    END,
    pay_url = COALESCE(pay_url, NULLIF(payment_url, '')),
    order_type = CASE
        WHEN COALESCE(biz_type, 'online_recharge') = 'subscription_purchase' THEN 'subscription'
        ELSE order_type
    END,
    subscription_group_id = COALESCE(subscription_group_id, biz_group_id),
    subscription_days = COALESCE(subscription_days, biz_validity_days),
    expires_at = COALESCE(expires_at, expire_at),
    completed_at = COALESCE(
        completed_at,
        CASE
            WHEN LOWER(COALESCE(status, '')) IN ('paid', 'completed', 'refunded') THEN paid_at
            ELSE NULL
        END
    );
