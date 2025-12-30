-- 用户活动状态与使用记录
CREATE TABLE IF NOT EXISTS user_promotions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    username        VARCHAR(100),

    activated_at    TIMESTAMPTZ NOT NULL,
    expire_at       TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'active',

    used_tier       INT,
    used_at         TIMESTAMPTZ,
    used_amount     DECIMAL(20, 8),
    bonus_amount    DECIMAL(20, 8),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_promotions_status_expire
    ON user_promotions(status, expire_at)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS promotion_records (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username    VARCHAR(100),
    tier        INT NOT NULL,
    amount      DECIMAL(20, 8) NOT NULL,
    bonus       DECIMAL(20, 8) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promotion_records_user_id
    ON promotion_records(user_id);

CREATE INDEX IF NOT EXISTS idx_promotion_records_created_at
    ON promotion_records(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_promotion_records_tier_created
    ON promotion_records(tier, created_at DESC);
