-- 203_invoice_tables.sql
-- 发票：开票申请、订单关联、用户默认开票信息（manual provider only）

CREATE TABLE IF NOT EXISTS invoice_requests (
    id                  BIGSERIAL PRIMARY KEY,
    invoice_request_no  VARCHAR(64) NOT NULL UNIQUE,
    user_id             BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status              VARCHAR(32) NOT NULL,

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

    amount_cny_total    DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_usd_total     DECIMAL(20, 8) NOT NULL DEFAULT 0,

    reviewed_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    reject_reason       TEXT NOT NULL DEFAULT '',
    issued_by           BIGINT REFERENCES users(id) ON DELETE SET NULL,
    issued_at           TIMESTAMPTZ,
    invoice_number      VARCHAR(64) NOT NULL DEFAULT '',
    invoice_date        DATE,
    invoice_pdf_url     TEXT NOT NULL DEFAULT '',

    provider            VARCHAR(50) NOT NULL DEFAULT 'manual',

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

    order_no            VARCHAR(64) NOT NULL,
    amount_cny          DECIMAL(20, 2) NOT NULL DEFAULT 0,
    total_usd           DECIMAL(20, 8) NOT NULL DEFAULT 0,
    active              BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_order_items_invoice_request_id
    ON invoice_order_items(invoice_request_id);

CREATE INDEX IF NOT EXISTS idx_invoice_order_items_payment_order_id
    ON invoice_order_items(payment_order_id);

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

INSERT INTO settings (key, value, updated_at)
VALUES ('invoice_default_item_name', '技术服务费', NOW())
ON CONFLICT (key) DO NOTHING;
