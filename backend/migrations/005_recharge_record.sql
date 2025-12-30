-- 充值/扣款流水记录表
CREATE TABLE IF NOT EXISTS recharge_records (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- 变动信息
    amount          DECIMAL(20, 8) NOT NULL,
    type            VARCHAR(20) NOT NULL,
    operator        VARCHAR(100),
    remark          VARCHAR(500),
    related_id      VARCHAR(100),

    -- 余额快照
    balance_before  DECIMAL(20, 8) NOT NULL,
    balance_after   DECIMAL(20, 8) NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recharge_records_user_created
    ON recharge_records(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recharge_records_type_created
    ON recharge_records(type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recharge_records_created_at
    ON recharge_records(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recharge_records_related_id
    ON recharge_records(related_id)
    WHERE related_id IS NOT NULL;
