# Sub2API User Lifecycle Analytics – API Notes

This skill uses only existing admin APIs:

## 1) Users export

- `GET /api/v1/admin/users/export`
  - Auth: `x-api-key: <admin-api-key>`
  - Returns CSV (all users, server-side pagination internally)

CSV columns (current implementation):

- `id`
- `email`
- `balance` (current USD balance)
- `created_at` (registration time, RFC3339)

## 2) Usage logs (detail)

- `GET /api/v1/admin/usage`
  - Auth: `x-api-key: <admin-api-key>`
  - Pagination: `page`, `page_size` (max 1000)
  - Date filters: `start_date`, `end_date` (YYYY-MM-DD) and `timezone`

Key JSON fields:

- `user_id`
- `model`
- `input_tokens`, `output_tokens` (cache tokens also exist)
- `actual_cost` (USD)
- `created_at` (RFC3339)

## 3) Recharge / payment orders export

- `GET /api/v1/admin/payment/orders/export`
  - Auth: `x-api-key: <admin-api-key>`
  - Filters: `status`, `provider`, `user_email`, `from`, `to`
    - `from/to` are RFC3339 and filter by order `created_at` in DB
  - Returns CSV

CSV columns (current implementation):

- `order_no`
- `order_type`
- `user_email` (used to join back to users)
- `provider`
- `amount_usd` (actual paid USD)
- `total_usd` (credited USD,到账额度)
- `status`
- `created_at`
- `paid_at`

## Join & metric semantics in this skill

- Join `payment/orders/export.user_email` -> `users/export.email` -> `user_id`
- “新增注册用户数/注册趋势/余额分布”:
  - Filter users by `users.created_at` within `--from..--to` in `--timezone`
  - Balance distribution uses current `balance` value
- “是否产生过 API 调用”:
  - Default: within the period (`/admin/usage` queried using `start_date/end_date`)
- “累计充值金额”:
  - Default: sum `total_usd` across paid orders (can switch to `amount_usd`)
- “累计消耗金额 / 最近一次消耗时间 / 首次消耗时间”:
  - Derived from `/admin/usage` by summing `actual_cost` and min/max of `created_at`

