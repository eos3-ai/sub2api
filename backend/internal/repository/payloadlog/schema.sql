CREATE TABLE IF NOT EXISTS {{database}}.payload_logs
(
    id                    UInt64,
    request_id            String,
    created_at            DateTime64(3, 'UTC'),
    username              LowCardinality(String),
    api_key_id            UInt64,
    api_key_name          String,
    account_id            UInt64,
    model_name            LowCardinality(String),
    method                LowCardinality(String),
    path                  LowCardinality(String),
    status_code           UInt16,
    is_stream             Bool,
    request_content_type  LowCardinality(String),
    response_content_type LowCardinality(String),
    request_bytes         UInt64,
    response_bytes        UInt64,
    request_body          String CODEC(ZSTD(3)),
    response_body         String CODEC(ZSTD(3)),
    INDEX idx_request_id request_id TYPE bloom_filter(0.001) GRANULARITY 1,
    INDEX idx_api_key_name api_key_name TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_username username TYPE set(1000) GRANULARITY 1
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(created_at)
ORDER BY (created_at, id)
SETTINGS index_granularity = 1024
