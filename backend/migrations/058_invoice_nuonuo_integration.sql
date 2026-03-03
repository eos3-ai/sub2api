-- 058_invoice_nuonuo_integration.sql
-- 为 invoice_requests 表添加诺诺开票集成相关字段

ALTER TABLE invoice_requests
  ADD COLUMN IF NOT EXISTS provider VARCHAR(50) NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS provider_invoice_id VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS provider_error TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS issuing_started_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_invoice_requests_provider
    ON invoice_requests(provider);

CREATE INDEX IF NOT EXISTS idx_invoice_requests_provider_invoice_id
    ON invoice_requests(provider_invoice_id)
    WHERE provider_invoice_id <> '';
