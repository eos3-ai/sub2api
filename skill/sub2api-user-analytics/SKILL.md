---
name: sub2api-user-analytics
description: Analyze Sub2API usage logs via the Admin API (x-api-key) to (1) compute period statistics (active users, per-model cost totals, per-model token shares, primary model distribution) and (2) export usage records to CSV (用户ID/调用时间/调用日期/模型/输入Tokens/输出Tokens/Tokens总量/实际消耗USD/调用状态). Use when asked to 统计使用记录、导出用量明细、统计模型分布与消耗、或需要基于 `/api/v1/admin/usage` 做报表。
---

# Sub2API 使用记录统计与导出（Usage Logs）

## Quick start

- 准备配置文件（包含 `url` + `admin_key`），示例：
  - `skill/sub2api-user-analytics/references/config.example.env`
  - `skill/sub2api-user-analytics/references/config.example.yaml`
- 统计某个周期内的用量与模型分布（默认 Tokens=输入+输出）：
  - `python3 skill/sub2api-user-analytics/scripts/usage_report.py summary --config /path/to/config.env --from 2026-01-01 --to 2026-01-31`
- 导出某个周期内的用量明细 CSV：
  - `python3 skill/sub2api-user-analytics/scripts/usage_report.py export --config /path/to/config.env --from 2026-01-01 --to 2026-01-31 --out ./usage_logs.csv`

## What this skill computes

### Period stats (summary)

- **有使用记录的用户数**：周期内 `user_id` 去重数量
- **主要使用的模型分布**：对每个 `user_id`，按 `--primary-by`（默认 tokens）选出该用户在周期内“占比最高”的模型，再统计各模型对应的用户数分布
- **各模型的总消耗金额（USD）**：周期内按 `model` 分组汇总 `actual_cost`
- **各模型 Tokens 消耗占比**：按 `model` 分组汇总 `tokens_total`，再除以全模型 `tokens_total` 总和

### Export fields (CSV)

- 用户ID -> `user_id`
- 调用时间 -> `created_at`（按 `--timezone` 转换并输出）
- 调用日期 -> 调用时间的日期部分
- 模型名称 -> `model`
- 输入 Tokens -> `input_tokens`
- 输出 Tokens -> `output_tokens`
- Tokens 总量 -> 默认 `input_tokens + output_tokens`（可用 `--include-cache-tokens` 把 cache tokens 也计入）
- 实际消耗金额（USD） -> `actual_cost`
- 调用状态（成功/失败） -> 由于 `/api/v1/admin/usage` 是“消耗记录”视角，当前导出默认统一标记为“成功”

## Notes / gotchas

- `/api/v1/admin/usage` 的过滤参数是 `start_date` / `end_date`（YYYY-MM-DD），本脚本用 `--from/--to` 直接映射到这两个参数。
- 如果你确实需要把失败请求也导出为“失败”，需要额外使用 Ops 监控相关接口（但失败请求不一定具备 tokens/cost 字段）。当前脚本按“使用消耗”口径导出。

## Reference

- API 参数与字段对照：`skill/sub2api-user-analytics/references/api.md`

