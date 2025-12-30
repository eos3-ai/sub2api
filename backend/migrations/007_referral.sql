-- 邀请码与邀请关系
CREATE TABLE IF NOT EXISTS referral_codes (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    code        VARCHAR(20) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS referral_invites (
    id                  BIGSERIAL PRIMARY KEY,

    invitee_id          BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    invitee_username    VARCHAR(100),
    referrer_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referrer_username   VARCHAR(100),

    total_recharge_usd  DECIMAL(20, 8) NOT NULL DEFAULT 0,

    is_qualified        BOOLEAN NOT NULL DEFAULT FALSE,
    qualified_at        TIMESTAMPTZ,
    reward_issued       BOOLEAN NOT NULL DEFAULT FALSE,
    reward_issued_at    TIMESTAMPTZ,
    reward_amount_usd   DECIMAL(20, 8) NOT NULL DEFAULT 0,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_referral_invites_referrer_created
    ON referral_invites(referrer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_referral_invites_pending_reward
    ON referral_invites(is_qualified, reward_issued)
    WHERE is_qualified = TRUE AND reward_issued = FALSE;

CREATE INDEX IF NOT EXISTS idx_referral_invites_referrer_stats
    ON referral_invites(referrer_id)
    INCLUDE (is_qualified, reward_issued, reward_amount_usd);
