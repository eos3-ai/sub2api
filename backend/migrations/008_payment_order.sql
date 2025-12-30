-- 支付订单表
CREATE TABLE IF NOT EXISTS payment_orders (
    id              BIGSERIAL PRIMARY KEY,
    order_no        VARCHAR(50) NOT NULL UNIQUE,
    trade_no        VARCHAR(100) UNIQUE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username        VARCHAR(100),

    amount_cny      DECIMAL(20, 2) NOT NULL,
    amount_usd      DECIMAL(20, 8) NOT NULL,
    bonus_usd       DECIMAL(20, 8) NOT NULL DEFAULT 0,
    total_usd       DECIMAL(20, 8) NOT NULL,
    exchange_rate   DECIMAL(10, 4) NOT NULL,

    provider        VARCHAR(20) NOT NULL,
    payment_method  VARCHAR(50),
    payment_url     VARCHAR(1000),

    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    paid_at         TIMESTAMPTZ,
    expire_at       TIMESTAMPTZ NOT NULL,

    promotion_tier  INT,
    promotion_used  BOOLEAN NOT NULL DEFAULT FALSE,

    callback_data   TEXT,
    callback_at     TIMESTAMPTZ,

    client_ip       VARCHAR(50),
    user_agent      VARCHAR(500),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_created
    ON payment_orders(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_status_created
    ON payment_orders(user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_created_at
    ON payment_orders(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_status_created
    ON payment_orders(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_provider_status
    ON payment_orders(provider, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_orders_pending_expire
    ON payment_orders(expire_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_payment_orders_paid_stats
    ON payment_orders(paid_at DESC)
    INCLUDE (amount_cny, amount_usd, bonus_usd, total_usd, provider)
    WHERE status = 'paid';
