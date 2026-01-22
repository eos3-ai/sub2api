# Sub2API Recharge (Payment Orders) API Notes

This skill reads recharge data from the existing admin endpoint:

- `GET /api/v1/admin/payment/orders`
  - Auth: `x-api-key: <admin-api-key>`
  - Pagination: `page`, `page_size` (max 1000)
  - Filters:
    - `status`: `paid|pending|failed|expired|cancelled|refunded`
    - `provider`: e.g. `zpay|stripe|admin|activity`
    - `method`: legacy alias; if set, backend maps `alipay -> zpay`, `wechat -> stripe`
    - `from` / `to` (RFC3339): filters by `created_at` in DB

Response envelope:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [ { /* PaymentOrder DTO */ } ],
    "total": 123,
    "page": 1,
    "page_size": 1000,
    "pages": 1
  }
}
```

Key DTO fields used by this skill:

- `order_no`: order identifier (string)
- `user_id`: user identifier (int)
- `created_at`: order creation time (RFC3339 string)
- `paid_at`: payment success time (RFC3339 string, nullable)
- `amount_cny`: paid CNY amount
- `amount_usd`: paid USD amount (actual payment)
- `total_usd`: credited USD amount (到账额度, usually includes bonus)
- `payment_method`: payment method label (string, may be empty)
- `channel` / `provider`: fallbacks for payment method display

CSV export requirements mapping (default behavior of `scripts/recharge_report.py export`):

- 订单ID -> `order_no`
- 用户ID -> `user_id`
- 充值时间 -> `paid_at` (fallback `created_at`)
- 充值日期 -> date part of the above (in `--timezone`)
- 实际支付金额（USD） -> `amount_usd`
- 到账额度（USD） -> `total_usd`
- 是否首充 -> computed (see `--first-scope`)
- 支付方式 -> `payment_method` (fallback `channel`, then `provider`)

