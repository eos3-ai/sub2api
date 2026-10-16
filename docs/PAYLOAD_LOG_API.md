# Payload Log API

Payload logging is disabled by default. Set `PAYLOAD_LOG_ENABLED=true`, start
the ClickHouse service, and then enable recording through the admin API. The
feature stores normalized `messages` only; request headers, generation options,
and upstream raw payloads are not stored.

All endpoints require an administrator JWT or a read/write administrator API
key (`x-api-key`). Ordinary user keys, sales users, and read-only admin keys do
not have access.

## Runtime settings

```http
GET /api/v1/admin/payload-logs/settings
PUT /api/v1/admin/payload-logs/settings
Content-Type: application/json

{"enabled":true}
```

The response contains `enabled`, `store_available`, `store_error`, and the
instance-local `dropped` counter. Recording starts only when the environment
feature gate, ClickHouse availability, and runtime switch are all enabled.

## List metadata

```http
GET /api/v1/admin/payload-logs?page=1&page_size=20&username=alice
```

Optional filters are `username`, `api_key_name` (or `token_name`),
`model_name`, `start_timestamp`, and `end_timestamp`. The default range is the
last 24 hours and the maximum range is 31 days. List responses never contain
message bodies.

## Get message bodies

```http
GET /api/v1/admin/payload-logs/{request_id}
```

The request ID is the value of `X-Client-Request-ID`; a `client:` prefix is
accepted for IDs copied from usage logs. Each `request_body` and
`response_body` field is a JSON string containing a `messages` array. Body
access is audited as `payload_log.view` without copying the message content
into the audit record.
