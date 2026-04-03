-- Add user-facing subscription purchase settings on groups.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS user_purchase_visible BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS user_purchase_price_usd DECIMAL(20,8);

COMMENT ON COLUMN groups.user_purchase_visible IS '是否在用户充值页展示为可购买套餐（仅订阅分组）';
COMMENT ON COLUMN groups.user_purchase_price_usd IS '用户侧套餐购买价格（USD），为空表示不可购买';

CREATE INDEX IF NOT EXISTS idx_groups_user_purchase_visible
    ON groups(user_purchase_visible)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_groups_subscription_user_purchase
    ON groups(subscription_type, user_purchase_visible)
    WHERE deleted_at IS NULL;

-- Add business fields to payment_orders for differentiating purchase intent.
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS biz_type VARCHAR(40) NOT NULL DEFAULT 'online_recharge';

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS biz_group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL;

ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS biz_validity_days INT;

COMMENT ON COLUMN payment_orders.biz_type IS '订单业务类型：online_recharge/subscription_purchase';
COMMENT ON COLUMN payment_orders.biz_group_id IS '订阅购买关联的分组 ID';
COMMENT ON COLUMN payment_orders.biz_validity_days IS '订阅购买有效期（天）';

UPDATE payment_orders
SET biz_type = 'online_recharge'
WHERE biz_type IS NULL OR biz_type = '';

CREATE INDEX IF NOT EXISTS idx_payment_orders_biz_type_created
    ON payment_orders(biz_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_biz_group_id
    ON payment_orders(biz_group_id)
    WHERE biz_group_id IS NOT NULL;
