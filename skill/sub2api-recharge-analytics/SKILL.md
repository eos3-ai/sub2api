---
name: sub2api-recharge-analytics
description: Analyze Sub2API recharge/payment orders via the Admin API (x-api-key) to (1) compute recharge stats for a time range and (2) export recharge order details to CSV (订单ID/用户ID/充值时间/充值日期/实际支付金额USD/到账额度USD/是否首充/支付方式). Use when asked to 统计充值数据、导出充值订单、对账支付订单、或需要基于 `/api/v1/admin/payment/orders` 做分析报表。
---

# Sub2API 充值数据统计与导出

## Quick start

- 准备一个配置文件（推荐 `.env` 或简单 YAML），示例见：
  - `skill/sub2api-recharge-analytics/references/config.example.env`
  - `skill/sub2api-recharge-analytics/references/config.example.yaml`
- 统计某个周期的充值数据（默认仅统计 `status=paid`）：
  - `python3 skill/sub2api-recharge-analytics/scripts/recharge_report.py summary --config /path/to/config.env --from 2026-01-01 --to 2026-01-31`
- 导出某个周期的充值订单明细到 CSV：
  - `python3 skill/sub2api-recharge-analytics/scripts/recharge_report.py export --config /path/to/config.env --from 2026-01-01 --to 2026-01-31 --out ./recharge_orders.csv --first-scope global`

## Inputs

- `--config`：包含 `url`（或 `base_url`）与 `admin_key`（或 `admin_api_key`）的配置文件
- `--from/--to`：
  - 支持 `YYYY-MM-DD`（按 `--timezone` 解释；`--to` 会取当天 23:59:59）
  - 或 RFC3339（如 `2026-01-01T00:00:00+08:00` / `2026-01-01T00:00:00Z`）
- 可选过滤：
  - `--provider` / `--method` / `--status`（对应后端的 query 参数）

## Notes / gotchas

- 后端接口的 `from/to` 过滤字段是 `created_at`（下单时间），不是 `paid_at`（支付成功时间）。导出字段里的“充值时间”优先用 `paid_at`，没有则回退到 `created_at`。
- “是否首充”支持两种算法（`--first-scope`）：
  - `global`：扫描历史所有已支付订单，判断该用户的最早支付订单是否为本单（更符合“首充”语义，但会多一次全量扫描）
  - `period`：仅在本周期内判断“首单”（更快，但不等价于全局首充）

## Reference

- 端点、鉴权、字段对照等：`skill/sub2api-recharge-analytics/references/api.md`

