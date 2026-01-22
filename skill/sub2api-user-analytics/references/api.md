# Sub2API Usage Logs API Notes

This skill reads usage/cost records from the existing admin endpoint:

- `GET /api/v1/admin/usage`
  - Auth: `x-api-key: <admin-api-key>` (read-only key typically allows this endpoint)
  - Pagination: `page`, `page_size` (max 1000)
  - Date filters:
    - `start_date` / `end_date` (YYYY-MM-DD)
    - `timezone` (IANA tz name, e.g. `Asia/Shanghai`), used by server to interpret date boundaries
  - Optional filters:
    - `user_id`, `api_key_id`, `account_id`, `group_id`
    - `model`
    - `stream` (`true|false`)
    - `billing_type` (`0|1`)

Response envelope:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [ { /* UsageLog DTO */ } ],
    "total": 123,
    "page": 1,
    "page_size": 1000,
    "pages": 1
  }
}
```

Key DTO fields used by this skill:

- `user_id`: user identifier
- `created_at`: call time (RFC3339 string)
- `model`: model name
- `input_tokens`, `output_tokens`
- `cache_creation_tokens`, `cache_read_tokens` (optional)
- `actual_cost`: user billed USD

CSV mapping (default behavior of `scripts/usage_report.py export`):

- 用户ID -> `user_id`
- 调用时间 -> `created_at` (converted to `--timezone`)
- 调用日期 -> date part of converted call time
- 模型名称 -> `model`
- 输入 Tokens -> `input_tokens`
- 输出 Tokens -> `output_tokens`
- Tokens 总量 -> `input_tokens + output_tokens` (or include cache tokens if `--include-cache-tokens`)
- 实际消耗金额（USD） -> `actual_cost`
- 调用状态 -> default `成功` (usage logs represent billed usage records)

