---
name: sub2api-user-lifecycle-analytics
description: Analyze Sub2API user growth and lifecycle via existing Admin APIs (x-api-key) using `/api/v1/admin/users/export`, `/api/v1/admin/usage`, and `/api/v1/admin/payment/orders/export` to (1) compute period registration/balance statistics and (2) export user lifecycle CSVs (registration→activation, balance+recharge+usage totals, first recharge/usage + period usage). Use when asked to 统计新增注册用户、注册趋势、余额分布、用户激活（是否调用过 API）、或导出用户充值/消耗生命周期报表。
---

# Sub2API 用户增长与生命周期分析

## Quick start

- 准备配置文件（包含 `url` + `admin_key`），示例：
  - `skill/sub2api-user-lifecycle-analytics/references/config.example.env`
  - `skill/sub2api-user-lifecycle-analytics/references/config.example.yaml`

### 1) 统计周期内注册/余额指标

`python3 skill/sub2api-user-lifecycle-analytics/scripts/user_lifecycle_report.py summary --config /path/to/config.env --from 2026-01-01 --to 2026-01-31`

### 2) 导出：注册 + 是否产生过 API 调用

`python3 skill/sub2api-user-lifecycle-analytics/scripts/user_lifecycle_report.py export-registration --config /path/to/config.env --from 2026-01-01 --to 2026-01-31 --out ./users_activation.csv`

### 3) 导出：余额 + 累计充值/消耗 + 最近一次消耗时间

`python3 skill/sub2api-user-lifecycle-analytics/scripts/user_lifecycle_report.py export-finance --config /path/to/config.env --from 2026-01-01 --to 2026-01-31 --out ./users_finance.csv`

### 4) 导出：注册/首充/首耗 + 本周期是否消耗 + 本周期消耗金额

`python3 skill/sub2api-user-lifecycle-analytics/scripts/user_lifecycle_report.py export-lifecycle --config /path/to/config.env --from 2026-01-01 --to 2026-01-31 --out ./users_lifecycle.csv`

## Notes / gotchas

- 用户列表来自 `GET /api/v1/admin/users/export`（CSV），注册时间字段为 `created_at`。
- 用量明细来自 `GET /api/v1/admin/usage`（分页 JSON）；本 Skill 以 `actual_cost` 作为“消耗金额（USD）”口径。
- 充值记录来自 `GET /api/v1/admin/payment/orders/export`（CSV）；脚本默认以 `total_usd`（到账额度/入账 USD）作为“累计充值金额”，可用 `--recharge-field amount_usd` 改为按实际支付 USD 汇总。
- 充值记录 CSV 里是 `user_email`，脚本通过 `users/export` 的 email->id 映射进行关联；无 email 的订单会被跳过。

## Reference

- 端点/字段/口径说明：`skill/sub2api-user-lifecycle-analytics/references/api.md`

