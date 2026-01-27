-- 044_invoice_tables.sql
-- 发票：开票申请、订单关联、用户默认开票信息

CREATE TABLE IF NOT EXISTS invoice_requests (
    id                  BIGSERIAL PRIMARY KEY,
    invoice_request_no  VARCHAR(64) NOT NULL UNIQUE,
    user_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status              VARCHAR(32) NOT NULL,

    -- 发票类型：normal(普票) / special(专票)
    invoice_type        VARCHAR(16) NOT NULL,
    -- 抬头类型：personal / company
    buyer_type          VARCHAR(16) NOT NULL,
    invoice_title       VARCHAR(255) NOT NULL,
    tax_no              VARCHAR(64) NOT NULL DEFAULT '',

    -- 专票必填（MVP 仅做后端校验，不做 DB 约束）
    buyer_address       VARCHAR(255) NOT NULL DEFAULT '',
    buyer_phone         VARCHAR(32) NOT NULL DEFAULT '',
    buyer_bank_name     VARCHAR(128) NOT NULL DEFAULT '',
    buyer_bank_account  VARCHAR(64) NOT NULL DEFAULT '',

    receiver_email      VARCHAR(255) NOT NULL,
    receiver_phone      VARCHAR(32) NOT NULL DEFAULT '',

    -- 开票内容（默认：技术服务费，可由后台设置覆盖）
    invoice_item_name   VARCHAR(255) NOT NULL DEFAULT '',
    -- 备注（用户自填，可选）
    remark              TEXT NOT NULL DEFAULT '',

    -- 金额快照
    amount_cny_total    DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_usd_total     DECIMAL(20, 8) NOT NULL DEFAULT 0,

    -- 审核/开票信息
    reviewed_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    reject_reason       TEXT NOT NULL DEFAULT '',
    issued_by           BIGINT REFERENCES users(id) ON DELETE SET NULL,
    issued_at           TIMESTAMPTZ,
    invoice_number      VARCHAR(64) NOT NULL DEFAULT '',
    invoice_date        DATE,
    invoice_pdf_url     TEXT NOT NULL DEFAULT '',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_requests_user_created_at
    ON invoice_requests(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_invoice_requests_status_created_at
    ON invoice_requests(status, created_at DESC);

CREATE TABLE IF NOT EXISTS invoice_order_items (
    id                  BIGSERIAL PRIMARY KEY,
    invoice_request_id  BIGINT NOT NULL REFERENCES invoice_requests(id) ON DELETE CASCADE,
    payment_order_id    BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE RESTRICT,

    -- 冗余字段，便于展示与审计（避免依赖 payment_orders 变更）
    order_no            VARCHAR(50) NOT NULL,
    amount_cny          DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_usd           DECIMAL(20, 8) NOT NULL DEFAULT 0,

    -- 订单是否仍被该申请“占用”。当申请被取消/驳回时置为 false，从而允许重新开票。
    active              BOOLEAN NOT NULL DEFAULT TRUE,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_order_items_invoice_request_id
    ON invoice_order_items(invoice_request_id);

CREATE INDEX IF NOT EXISTS idx_invoice_order_items_payment_order_id
    ON invoice_order_items(payment_order_id);

-- 防重复开票（仅对 active=true 的记录生效；取消/驳回会释放占用）
CREATE UNIQUE INDEX IF NOT EXISTS ux_invoice_order_items_payment_order_active
    ON invoice_order_items(payment_order_id)
    WHERE active;

CREATE TABLE IF NOT EXISTS invoice_profiles (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    invoice_type        VARCHAR(16) NOT NULL,
    buyer_type          VARCHAR(16) NOT NULL,
    invoice_title       VARCHAR(255) NOT NULL,
    tax_no              VARCHAR(64) NOT NULL DEFAULT '',
    buyer_address       VARCHAR(255) NOT NULL DEFAULT '',
    buyer_phone         VARCHAR(32) NOT NULL DEFAULT '',
    buyer_bank_name     VARCHAR(128) NOT NULL DEFAULT '',
    buyer_bank_account  VARCHAR(64) NOT NULL DEFAULT '',
    receiver_email      VARCHAR(255) NOT NULL,
    receiver_phone      VARCHAR(32) NOT NULL DEFAULT '',
    invoice_item_name   VARCHAR(255) NOT NULL DEFAULT '',
    remark              TEXT NOT NULL DEFAULT '',

    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

