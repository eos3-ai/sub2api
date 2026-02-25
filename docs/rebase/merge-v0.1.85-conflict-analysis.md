# merge `v0.1.85` 到 `zyp-dev` 冲突分析与解决建议

- 日期：2026-02-24
- 合并方式：`git merge --no-ff --no-commit v0.1.85`
- 基线分支：`zyp-dev`
- 冲突文件数：28
- 冲突块数：54

---

## 1. 结论先行：为什么这次会冲突

这批冲突不是“单点改动撞车”，而是**两条分支在同一时期做了并行架构演进**，尤其集中在：

1. **后端网关与调度能力增强**（OpenAI/Gateway/Sora/并发与超时处理）。
2. **配置体系扩展**（`config.go`、`deploy/.env.example`、`deploy/config.example.yaml` 同步引入大量新字段）。
3. **DI/Wire 注入链路变化**（`wire.go`/`wire_gen.go` 参数与 provider 集变化）。
4. **运维与日志体系升级**（结构化日志、ops 系统日志、清理任务）。
5. **前端管理面模型/账号能力扩展**（Sora、更多筛选项、Codex 相关字段）。

---

## 2. 冲突热点（优先处理）

按冲突块数量排序：

| 文件 | 冲突块 | 主要原因 |
|---|---:|---|
| `backend/cmd/server/wire_gen.go` | 6 | 自动生成代码与构造函数签名并行变化 |
| `backend/internal/config/config.go` | 6 | 配置结构、加载流程、校验逻辑同时扩展 |
| `backend/internal/service/admin_service.go` | 5 | Admin 服务职责扩展方向不同（Sora/余额/告警） |
| `backend/internal/handler/openai_gateway_handler.go` | 3 | 并发控制与计费错误处理策略增强 |
| `backend/internal/server/middleware/admin_auth_test.go` | 3 | add/add 型测试文件并行新增 |
| `deploy/.env.example` | 3 | 环境变量注释与参数体系并行扩展 |
| `deploy/config.example.yaml` | 3 | 网关/Sora/Codex 配置项并行扩展 |

---

## 3. 分模块冲突原因与解决建议

### 3.1 构建与依赖

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `.gitignore` | 双方都在尾部追加忽略规则（`docs/*`、`.codex`、`frontend/coverage/`、`.remote-verify/` 等） | **手工并集**，但谨慎 `docs/*`（会屏蔽后续文档变更）；建议仅忽略生成物子目录 |
| `Makefile` | `npm` 与 `pnpm` 命令分歧，同时 v0.1.85 增加 `secret-scan` 目标 | 建议保留 `secret-scan`，前端命令统一到团队实际包管理器（推荐 `pnpm`） |
| `backend/cmd/server/VERSION` | 版本号冲突（`0.1.76` vs `0.1.85`） | 按发布策略取值；若本次目标是吸收 tag 能力，建议不低于 `0.1.85` |
| `backend/go.mod` | v0.1.85 新增 `github.com/alitto/pond/v2` 依赖（worker pool） | 建议保留该依赖并执行 `go mod tidy` |

### 3.2 DI / Wire（核心）

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/cmd/server/wire.go` | 清理流程和服务注入链变更：`PaymentMaintenanceService` 与 `OpsSystemLogSink`/`SoraMediaCleanupService`/更多 stop 步骤并行 | **语义合并**，清理步骤全部保留，确保 stop 顺序合理 |
| `backend/cmd/server/wire_gen.go` | 构造函数签名大面积变化：`AuthService`、`UserService`、`SubscriptionService`、`AdminService`、`ProvideHandlers` 参数均变化 | 不建议手工硬拼，先合并 `wire.go` 与 service 构造函数，再**重新生成 `wire_gen.go`** |
| `backend/internal/handler/wire.go` | `NewDingtalkBotHandler` 与 `NewSoraGatewayHandler` 同位置冲突 | 两者都应保留（分别对应 zyp-dev 与 v0.1.85 功能） |

### 3.3 配置体系（核心）

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/config/config.go` | `Config` 结构新增 `log`、`gateway.usage_record` 等；加载入口从 `Load` 扩展到 `load(...)`；校验逻辑从“release 限制”扩展到“全局 jwt/log 严格校验” | 采用“骨架取 v0.1.85 + 回填 zyp-dev 特性”：保留 `Log`/新校验/`LoadForBootstrap`，并补回 `AdminAPIKeyReadOnly`、`AnthropicAPIKeyMonitor`、本地安全约束；同步更新测试 |
| `deploy/.env.example` | DB/Redis/Gateway 参数注释体系重写 + 新增 Codex/Gateway 调度参数 | 逐段合并，避免重复键（如 `DATABASE_PORT`），保留新参数并校正默认值一致性 |
| `deploy/config.example.yaml` | 网关连接池默认值、Sora 参数、Codex 参数并行演进 | 以“字段齐全”为主，默认值按生产稳态审定（特别是 `max_conns_per_host`、`max_idle_conns`） |

### 3.4 Handler / Service 业务逻辑

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/handler/openai_gateway_handler.go` | v0.1.85 引入用户槽位快速路径、等待计数优化、账单错误映射、Ops 时延埋点；zyp-dev 保留既有路径 | 保留 v0.1.85 并发与埋点逻辑，同时确保 zyp-dev 自定义错误处理不丢失 |
| `backend/internal/service/openai_gateway_service.go` | 流超时处理：`log.Printf` vs `logger.LegacyPrintf + HandleStreamTimeout` | 建议保留 `HandleStreamTimeout`（行为增强）+ 统一日志风格 |
| `backend/internal/service/token_refresh_service.go` | 错误路径中“告警通知”与“结构化日志”冲突 | 两者合并：既记录结构化日志，也保留账号告警通知链路 |
| `backend/internal/service/admin_service.go` | `soraAccountRepo`、Sora 定价字段、余额调整台账逻辑、告警/日志体系并行修改 | 采用“能力并集”：Sora 相关字段 + 余额台账一致性 + 告警/日志统一；构造函数参数保持单一真源 |
| `backend/internal/service/account_test_service.go` | 新增 Sora 账号测试路径，同时 Claude 测试函数签名变化 | 保留 Sora 分支并统一 `testClaudeAccountConnection` 签名调用 |
| `backend/internal/service/gemini_oauth_service.go` | Google One 流程新增 `project_id` 自动探测 | 建议保留 v0.1.85 逻辑（功能更完整） |
| `backend/internal/service/auth_service.go` | 仅日志实现冲突（`log.Printf` vs `logger.LegacyPrintf`） | 统一项目日志风格，建议 `logger.LegacyPrintf` |
| `backend/internal/handler/admin/user_handler.go` | import 区冲突（导出 CSV 相关 import 与 context import 并行） | 合并 import 后以编译器清理未使用项 |
| `backend/internal/setup/setup.go` | JWT 自动生成提示、release 模式密钥强校验、自动 admin 密码生成逻辑冲突 | 建议保留 release 强校验 + 自动 admin 密码能力，并统一 logger 输出 |

### 3.5 测试冲突

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/server/middleware/admin_auth_test.go` | add/add：一侧是只读 Admin Key 权限测试，一侧是 JWT token version 测试（含 build tag） | 建议拆分为两个测试文件并都保留，避免巨型冲突块反复回归 |
| `backend/internal/pkg/geminicli/oauth_test.go` | 文案+期望 scope 变化（Google One 自定义客户端场景） | 以当前实现为准校正断言；若实现走 Code Assist 作用域，保留 `DefaultCodeAssistScopes` |

### 3.6 前端冲突

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `frontend/src/components/account/AccountUsageCell.vue` | 新增 `WindowStats`/`resolveCodexUsageWindow` 引入 | 保留 v0.1.85 导入并检查未使用符号 |
| `frontend/src/components/account/EditAccountModal.vue` | 模型限制 UI 与 OpenAI 透传禁用提示并行改动 | 两者合并：保留原模式切换 + passthrough 禁用提示 |
| `frontend/src/components/admin/account/AccountTableFilters.vue` | 过滤项从 platform+status 扩展到 type/group/status(含 rate_limited) + Sora/Antigravity | 保留扩展过滤能力，同时按现有布局适配 |
| `frontend/src/components/common/GroupBadge.vue` | 新增 Sora 平台主题色 | 保留 Sora 色板并保持其他平台一致性 |
| `frontend/src/components/layout/AuthLayout.vue` | Logo/标题渲染改为 `settingsLoaded` 守卫 | 建议保留守卫，避免配置加载前闪烁 |
| `frontend/src/i18n/locales/en.ts` | `exportRecords` 与 `searchUsers` 文案冲突 | 两者都保留（新增键 + 新搜索文案） |
| `frontend/src/types/index.ts` | `user_agent`、`cache_ttl_overridden` 字段插入冲突，且存在重复可选字段 | 合并为单一定义，消除重复字段 |

---

## 4. 推荐解冲顺序（按风险）

1. **先低风险文件**：`.gitignore`、`Makefile`、`VERSION`、`go.mod`。
2. **再处理配置骨架**：`config.go`、`deploy/.env.example`、`deploy/config.example.yaml`。
3. **处理服务与 handler 核心冲突**：`admin_service`、`openai_gateway_handler`、`openai_gateway_service`、`token_refresh_service`。
4. **最后处理 Wire 与生成文件**：`wire.go` → `wire_gen.go`（重生成）。
5. **收尾前端与测试**：类型、组件、i18n、测试断言。

---

## 5. 建议的验收清单

建议每个阶段完成后做最小验证，避免一次性积累风险：

```bash
# 后端
go test ./backend/internal/service/... ./backend/internal/handler/... ./backend/internal/config/...

# 依赖与生成（按仓库实际工具调整）
go mod tidy

# 前端
pnpm --dir frontend run typecheck
pnpm --dir frontend run lint:check
```

如果 `wire_gen.go` 由工具生成，请在最终阶段统一生成并再跑一轮测试。

---

## 6. 本次结论（merge 视角）

- 对于当前仓库，`merge` 的冲突面是**一次性 28 文件**，相较 `rebase` 的多轮停顿更易做集中治理。
- 但本次冲突属于”能力并集型冲突”，不能简单 `accept ours/theirs`；核心文件应以**语义合并**为主。
- 成功关键点是：先稳住配置与依赖，再合并服务行为，最后重建生成代码与测试闭环。

---

## 7. 补充分析：遗漏冲突点与场景（基于全量 rebase 扫描对比）

> 本节通过对比 `rebase-v0.1.85-all-conflict-points.md`（67 个冲突文件）与原分析（28 个文件），
> 补充在 merge 场景下同样需要关注但原文未覆盖的冲突点。

---

### 7.1 `antigravity_gateway_service.go`（遗漏的核心冲突）

该文件是 rebase 扫描中**第一个**发生冲突的文件，但原分析完全未提及。

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/service/antigravity_gateway_service.go` | zyp-dev 已实现 `MODEL_CAPACITY_EXHAUSTED` 专项处理（不切换账号、直接返回上游）；v0.1.85 在相邻位置引入”单账号 503 退避”分支（`isSingleAccountRetry`），同时将 `rateLimitDuration` 的定义位置提前，导致变量作用域冲突 | 保留两类”提前返回”语义并排序：① 先判断 `MODEL_CAPACITY_EXHAUSTED`；② 再判断单账号 503；③ 最后走 `rateLimitDuration` 计算 + 限流 + 切换账号。变量 `rateLimitDuration` 定义移至两段提前返回之后 |

---

### 7.2 Sora 子系统（全新文件族 + 集成冲突）

v0.1.85 引入完整的 Sora 网关子系统，在 zyp-dev 中相关文件**均不存在**，属于”新增型”冲突（add/add）。需要接受全部新文件，同时处理集成点处与既有代码的冲突。

**7.2.1 Sora 核心实现（全新文件，直接接受）**

| 文件 | 说明 |
|---|---|
| `backend/internal/service/sora_client.go` | Sora API 客户端（令牌刷新、挑战处理、流式请求） |
| `backend/internal/service/sora_gateway_service.go` | Sora 网关调度服务（账号池管理、重试逻辑） |
| `backend/internal/handler/sora_gateway_handler.go` | Sora HTTP 路由入口处理器 |
| `backend/internal/service/sora_media_storage.go` | Sora 媒体资源存储与下载管理 |
| `backend/internal/service/sora_request_guard.go` | Sora 请求防护（直连安全、下载限制） |
| `deploy/docker-compose-aicodex.yml` | Sora/Codex 专项部署编排文件 |

这些文件在 zyp-dev 中不存在，merge 时应直接 `accept theirs`，无需手工合并。

**7.2.2 Sora 集成点冲突（需手工合并）**

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/handler/handler.go` | v0.1.85 注册 `NewSoraGatewayHandler`；zyp-dev 可能在同区域注册了其他 handler | 合并 handler 注册列表，两者均保留 |
| `backend/internal/server/routes/admin.go` | v0.1.85 新增 Sora 路由分组（`/sora/*`）；zyp-dev 在同文件扩展了管理路由 | 将 Sora 路由组追加到现有路由后，确保路径不冲突 |
| `backend/internal/handler/admin/group_handler.go` | v0.1.85 新增 `sora` 作为合法平台枚举值；zyp-dev 在相邻位置扩展了 antigravity 配置字段 | 枚举中同时保留 `sora`；确保现有平台字段不被覆盖 |
| `frontend/src/components/common/PlatformIcon.vue` | v0.1.85 为 Sora 添加 SVG 图标主题色；zyp-dev 仅支持四平台（anthropic/openai/gemini/antigravity） | 追加 Sora 色板与 SVG 路径，保持其他平台不变 |
| `frontend/src/views/admin/GroupsView.vue` | v0.1.85 在平台过滤选项中加入 `sora`；zyp-dev 在分组管理视图中有独立扩展（2264 行） | 仅在 platform 下拉/过滤条件中追加 sora，不改动 zyp-dev 的扩展功能 |
| `frontend/src/composables/useModelWhitelist.ts` | v0.1.85 在白名单中补充 Sora 模型条目；zyp-dev 已含 Claude opus-4.6 等最新模型 | 追加 Sora 模型条目，与 zyp-dev 最新模型并存 |
| `frontend/src/api/admin/index.ts` | v0.1.85 导出 Sora 相关 API 模块；zyp-dev 已有 12+ 模块导出 | 追加 Sora API 导出，避免覆盖现有导出 |

---

### 7.3 Ent ORM 生成文件（Schema 变更与重生成）

v0.1.85 引入 Sora 账号实体（新字段/新 Schema），导致 Ent 自动生成文件发生变化。zyp-dev 当前的 `mutation.go`（23,628 行）与 `runtime.go`（1,045 行）均不含任何 Sora 相关内容。

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/ent/mutation.go` | v0.1.85 在 schema 中新增 Sora 账号字段/关联，触发 Ent 代码重新生成 | **不要手工合并**；在所有 `backend/ent/schema/*.go` 变更合并完成后，执行 `go generate ./backend/ent/...` 重新生成，以生成文件为最终版本 |
| `backend/ent/runtime/runtime.go` | 同上，Sora schema 默认值与校验器初始化新增 | 同上，随 `mutation.go` 一起重新生成 |

> **重要**：如果 zyp-dev 侧也有自定义 Schema 变更（如公告、用户属性等），需先将 `backend/ent/schema/` 下的源文件手工合并，再统一重新生成，避免生成文件产生冲突。

---

### 7.4 遗漏的后端 Service / Middleware 文件

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/service/wire.go`（服务层 DI） | 与 `backend/cmd/server/wire.go` 为**不同文件**；v0.1.85 在服务层 wire 中添加 `SoraGatewayService`、`SoraMediaCleanupService` 等 provider；zyp-dev 在同文件也有扩展 | 合并 Provider 列表，建议以”功能并集”原则处理；注意 provider 函数签名已变化，与 `wire_gen.go` 重生成联动 |
| `backend/internal/service/gateway_request.go` | v0.1.85 优化重试场景下的 thinking 过滤性能（590 行）；zyp-dev 也在此文件有 ParsedRequest/SessionContext 扩展 | 保留双方优化：zyp-dev 的 SessionContext 扩展 + v0.1.85 的 thinking 过滤性能改进 |
| `backend/internal/service/subscription_service.go` | v0.1.85 修复批量更新凭证明细与缓存 TTL 抖动问题；zyp-dev 在此文件有 MaxExpiresAt/MaxValidityDays 约束扩展 | 保留 zyp-dev 约束字段；叠加 v0.1.85 的 TTL 修复 |
| `backend/internal/service/token_refresher.go` | v0.1.85 在 token 刷新器中引入与 Sora 相关的刷新策略扩展；zyp-dev 可能有告警/日志路径修改（与 `token_refresh_service.go` 联动） | 与 `token_refresh_service.go` 一起处理，保持刷新器与服务层行为一致 |
| `backend/internal/service/api_key_auth_cache.go` / `api_key_auth_cache_impl.go` | v0.1.85 在 API Key 认证缓存快照中可能新增 Sora 相关字段（如 sora_quota）；zyp-dev 已有 IP 白黑名单、负缓存等字段 | 合并快照结构，避免字段重复；确保缓存实现类与接口同步 |
| `backend/internal/server/middleware/cors.go` | v0.1.85 的 ETag 增量同步提交（`06b0f62e`）修改了 CORS 暴露头处理；zyp-dev 有 AllowCredentials 与通配符冲突处理逻辑 | 合并 CORS 暴露头列表，确保 ETag 相关头（如 `ETag`、`If-None-Match`）被正确暴露 |

---

### 7.5 遗漏的前端文件

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `frontend/src/i18n/locales/zh.ts` | 与 `en.ts` 同步，v0.1.85 在中文文案中添加 Sora 相关翻译；zyp-dev 在此文件有公告、用户侧功能等大量本地化扩展（4215 行） | 策略与 `en.ts` 一致：逐键合并，保留双方新增键；重点检查 `sora`、`exportRecords`、`searchUsers` 等键的中英文对应 |
| `frontend/src/components/account/CreateAccountModal.vue` | 与 `EditAccountModal.vue` 联动：v0.1.85 新增 Sora 平台账号创建 UI；zyp-dev 在创建模态框中有用户侧功能入口控制 | 两者均保留；复用 `EditAccountModal.vue` 的 passthrough/模式切换 UI 处理思路 |

---

### 7.6 遗漏的测试文件

除原分析中提及的 `admin_auth_test.go` 和 `oauth_test.go` 外，以下测试文件同样需要处理：

**后端测试：**

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `backend/internal/config/config_test.go` | v0.1.85 新增 `server.frontend_url` 安全校验测试用例；zyp-dev 也在此文件有配置校验测试 | 合并用例，保留双方断言；确保 `frontend_url` 禁止 query/userinfo 的校验在测试中有覆盖 |
| `backend/internal/service/billing_service_test.go` | v0.1.85 补充计费服务单测（覆盖率提升至 85%+）；zyp-dev 的计费逻辑有自定义扩展 | 以 v0.1.85 的覆盖率测试为基础，补充 zyp-dev 自定义计费路径的测试用例 |
| `backend/internal/service/openai_codex_transform_test.go` | v0.1.85 清理了无用的 opencode 缓存测试；zyp-dev 中可能保留了相关 mock | 以 v0.1.85 清理结果为准，确认 zyp-dev 中无对应缓存逻辑残留 |
| `backend/internal/service/openai_gateway_service_codex_cli_only_test.go` | v0.1.85 修复”拒绝日志记录原始 User-Agent”后更新了测试断言；zyp-dev 中 User-Agent 透传测试路径不同 | 对齐测试断言与实现：若实现走脱敏路径，则断言不含原始 UA |
| `backend/internal/service/admin_service_proxy_quality_test.go` | v0.1.85 将 401/405 质量检测结果调整为告警级别；zyp-dev 的代理质量检测逻辑有扩展 | 确保告警级别调整在测试中有对应断言；与 `admin_service.go` 合并保持联动 |
| `backend/internal/service/sora_media_cleanup_service_test.go` | v0.1.85 新增 Sora 媒体清理服务测试；zyp-dev 无此文件 | 直接接受（新增文件），确保测试中引用的依赖（数据库/对象存储）配置与 zyp-dev 一致 |
| `backend/internal/service/subscription_maintenance_queue_test.go` | v0.1.85 补充订阅维护队列单测；zyp-dev 在订阅服务有扩展 | 检查 zyp-dev 自定义约束（MaxExpiresAt 等）是否被测试覆盖，必要时补充用例 |
| `backend/internal/repository/idempotency_repo_integration_test.go` | v0.1.85 修复幂等测试中哈希值超 VARCHAR(64) 问题；zyp-dev 可能使用原始字符串 | 统一使用哈希值，避免字段长度限制导致集成测试在生产数据库中失败 |

**前端测试：**

| 文件 | 冲突原因 | 建议 |
|---|---|---|
| `frontend/src/api/__tests__/client.spec.ts` 等 6 个前端测试文件 | v0.1.85 的 `29ca1290`（清理测试用例与类型导入）批量修改了前端测试断言与快照；zyp-dev 在同测试文件中有公告、用户侧等相关用例 | 以”当前实现为准”校正断言；若使用 Vitest 快照，需重录（`vitest -u`） |

---

### 7.7 修订后的推荐解冲顺序

在原第 4 节基础上，补充以下调整：

1. **低风险文件**：`.gitignore`、`Makefile`、`VERSION`、`go.mod`（同原顺序）。
2. **配置骨架**：`config.go`、`deploy/.env.example`、`deploy/config.example.yaml`（同原顺序）。
3. **新增 Sora 核心文件**：直接 accept `sora_client.go`、`sora_gateway_service.go`、`sora_gateway_handler.go`、`sora_media_storage.go`、`sora_request_guard.go`、`docker-compose-aicodex.yml`。
4. **Sora 集成点与遗漏后端文件**：`antigravity_gateway_service.go`、`handler.go`、`routes/admin.go`、`group_handler.go`、`cors.go`、`gateway_request.go`、`subscription_service.go`、`token_refresher.go`、`api_key_auth_cache*`。
5. **原第 3 步**：`admin_service`、`openai_gateway_handler`、`openai_gateway_service`、`token_refresh_service`。
6. **Ent Schema + 重生成**：合并 `backend/ent/schema/*.go` 变更 → 执行 `go generate ./backend/ent/...`。
7. **Wire DI**：`backend/internal/service/wire.go` → `backend/cmd/server/wire.go` → 重生成 `wire_gen.go`。
8. **前端与 i18n**：`types/index.ts`、`PlatformIcon.vue`、`GroupsView.vue`、`AccountTableFilters.vue`、`EditAccountModal.vue`、`CreateAccountModal.vue`、`en.ts`、`zh.ts`、`useModelWhitelist.ts`、`admin/index.ts`。
9. **测试文件收尾**：按各模块对应实现，逐一校正断言或重录快照。

---

### 7.8 修订后的验收清单

```bash
# 后端单测（含新增 Sora 服务测试）
go test ./backend/internal/service/... \
        ./backend/internal/handler/... \
        ./backend/internal/config/... \
        ./backend/internal/repository/...

# Ent 代码重生成（Sora Schema 合并后执行）
go generate ./backend/ent/...

# Wire 重生成（wire.go 合并后执行）
go generate ./backend/cmd/server/...

# 依赖整理
go mod tidy

# 前端类型检查与 lint
pnpm --dir frontend run typecheck
pnpm --dir frontend run lint:check

# 前端测试（快照更新）
pnpm --dir frontend run test -- --run
# 若快照需更新：
pnpm --dir frontend run test -- --run -u
```

---

## 8. zyp-dev 功能保护清单（合并约束：不得影响已有功能）

> **核心约束**：本次 merge 必须以 zyp-dev 为主干，v0.1.85 的新能力只能以"追加/补充"方式集成，
> 不得覆盖、删除或重构任何 zyp-dev 已有功能。

---

### 8.1 zyp-dev 独有功能模块一览

以下功能模块在 v0.1.85 中**完全不存在**，合并时必须 100% 保留：

| 模块 | 后端关键文件 | 前端关键文件 |
|---|---|---|
| **支付系统** | `payment_handler.go`、`payment_order_repo.go`、`PaymentConfig` | `PaymentView.vue`、`PaymentOrdersView.vue` |
| **发票系统** | `invoice_handler.go`、`invoice_repo.go`、admin `InvoiceHandler` | `InvoicesView.vue`（用户/管理）、`InvoiceRequestModal.vue` |
| **推广活动** | `promotion_handler.go`、`promotion_repo.go`、`PromotionConfig` | `PromotionView.vue` |
| **邀请返利** | `referral_handler.go`、`referral_repo.go`、`ReferralConfig` | `ReferralView.vue` |
| **钉钉机器人** | `dingtalk_bot_handler.go`、`DingtalkBotConfig`、`routes/dingtalk.go` | — |
| **只读 Admin Key** | `middleware/admin_auth.go`（ReadOnly 逻辑）、`AdminAPIKeyReadOnlyConfig` | — |
| **Anthropic 账号监控** | `AnthropicAPIKeyMonitorService`、`AnthropicAPIKeyMonitorConfig` | — |
| **账号告警** | `AccountAlertService`（在 `admin_service.go`） | — |
| **公告系统** | `announcement` 相关 schema/repo | `AnnouncementsView.vue`、`UserDashboardAnnouncements.vue`、`useAnnouncementAutoPopup.ts` |
| **账单导出** | — | `billing.exportCsv/Excel` i18n + 导出 API |

---

### 8.2 高风险冲突文件：必须保留的 zyp-dev 内容

以下文件在 merge 时发生冲突，且 zyp-dev 侧有大量独有内容，**如果错误地 `accept theirs`，将直接导致功能丢失**。

---

#### `backend/internal/handler/handler.go`

**风险**：v0.1.85 新增 SoraGatewayHandler，若直接取 theirs 将丢失 zyp-dev 所有业务 handler。

**必须保留（zyp-dev 侧）**：
```
Handlers.Payment       *PaymentHandler
Handlers.Invoice       *InvoiceHandler
Handlers.Promotion     *PromotionHandler
Handlers.Referral      *ReferralHandler
Handlers.DingtalkBot   *DingtalkBotHandler
AdminHandlers.PaymentOrders  *admin.PaymentOrdersHandler
AdminHandlers.Invoices       *admin.InvoiceHandler
```

**操作建议**：取 zyp-dev 版本为基础，仅追加 v0.1.85 的 `SoraGatewayHandler` 字段。

---

#### `backend/internal/handler/wire.go`

**风险**：v0.1.85 新增 `NewSoraGatewayHandler`，若取 theirs 将丢失 zyp-dev 注册的所有 handler provider。

**必须保留（zyp-dev 侧）**：
```
NewPaymentHandler
NewInvoiceHandler
NewPromotionHandler
NewReferralHandler
NewDingtalkBotHandler
admin.NewPaymentOrdersHandler
admin.NewInvoiceHandler
```

**操作建议**：取 zyp-dev 版本为基础，追加 `NewSoraGatewayHandler`。

---

#### `backend/internal/server/routes/admin.go`

**风险**：v0.1.85 新增 Sora 路由，若取 theirs 将丢失支付订单和发票管理路由。

**必须保留（zyp-dev 侧）**：
```
registerPaymentOrderRoutes(admin, h)
  GET  /admin/payment/orders
  GET  /admin/payment/orders/summary
  GET  /admin/payment/orders/export

registerInvoiceRoutes(admin, h)
  GET  /admin/invoices
  GET  /admin/invoices/export
  GET  /admin/invoices/:id
  POST /admin/invoices/:id/approve
  POST /admin/invoices/:id/reject
  POST /admin/invoices/:id/issue
```

**操作建议**：取 zyp-dev 版本为基础，追加 Sora 路由注册（`/sora/*`）。

---

#### `backend/internal/config/config.go`

**风险**：v0.1.85 对 `Config` 结构体做了大量扩展（Log、gateway.usage_record 等），若采用 v0.1.85 骨架而不回填 zyp-dev 字段，将导致启动时配置反序列化失败。

**必须保留（zyp-dev 侧）**：
```go
// Config 结构体字段
Promotion    PromotionConfig
Referral     ReferralConfig
Payment      PaymentConfig
Dingtalk     DingtalkConfig
DingtalkBot  DingtalkBotConfig

// Security 节
AdminAPIKeyReadOnly  AdminAPIKeyReadOnlyConfig

// Gateway.Scheduling 节
AnthropicAPIKeyMonitor  AnthropicAPIKeyMonitorConfig

// 完整类型定义
PromotionConfig / PromotionTier
ReferralConfig
PaymentConfig / PaymentPackage / ZpayConfig / StripeConfig
DingtalkConfig / DingtalkBotConfig
AdminAPIKeyReadOnlyConfig
AnthropicAPIKeyMonitorConfig
```

**操作建议**：以"字段并集"原则合并，v0.1.85 的 `Log`/`LoadForBootstrap`/严格校验保留，zyp-dev 的 12+ 个配置字段同步保留。合并后需跑配置加载测试确保无字段缺失。

---

#### `backend/cmd/server/wire.go`

**风险**：v0.1.85 在 wire 中新增 `OpsSystemLogSink`、`SoraMediaCleanupService` 等，若取 theirs 将丢失 zyp-dev 的 cleanup 服务注入。

**必须保留（zyp-dev 侧）**：
```go
// provideCleanup 中的服务
paymentMaintenance        *service.PaymentMaintenanceService
anthropicAPIKeyMonitor    *service.AnthropicAPIKeyMonitorService
```

**操作建议**：清理步骤全部并集保留，确保 zyp-dev 的支付维护和账号监控服务在关闭时正常 stop。

---

#### `backend/internal/server/middleware/admin_auth.go`

**风险**：此文件在 v0.1.85 中不存在对应修改，但若因 merge 导致文件被重置为旧版本，将丢失只读 Admin Key 全部逻辑。

**必须保留（zyp-dev 侧）**：
```go
type adminAPIKeyAccess int  // readWrite / readOnly 枚举
type readOnlyAdminAPIKeyAllowlist struct { exact map; prefixes []string }

func isReadOnlyAdminAPIKeyMethod(method string) bool
func matchAdminAPIKeyAccess(...) (adminAPIKeyAccess, bool, error)
func newReadOnlyAdminAPIKeyAllowlist(cfg *config.Config) readOnlyAdminAPIKeyAllowlist
func (a readOnlyAdminAPIKeyAllowlist) Allows(rawPath string) bool

// 默认只读路径白名单
"/api/v1/admin/users/export"
"/api/v1/admin/usage"
"/api/v1/admin/payment/orders/summary"
"/api/v1/admin/payment/orders/export"
"/api/v1/admin/invoices/export"
```

**操作建议**：该文件无冲突时直接保留 zyp-dev 版本；如有冲突，只允许追加 v0.1.85 的逻辑，不得修改已有判断路径。

---

#### `backend/internal/service/admin_service.go`

**风险**：v0.1.85 增加 Sora 定价字段和余额台账逻辑，若取 theirs 将丢失 zyp-dev 的账号告警服务。

**必须保留（zyp-dev 侧）**：
```go
accountAlert *AccountAlertService  // 告警服务依赖注入
// 以及所有调用 accountAlert 的告警路径
```

**操作建议**：构造函数参数并集：保留 zyp-dev 的 `accountAlert`，同时添加 v0.1.85 的 Sora 相关参数。

---

#### `frontend/src/types/index.ts`

**风险**：v0.1.85 新增 Sora/Codex 类型，若取 theirs 将丢失 zyp-dev 的业务类型定义。

**必须保留（zyp-dev 侧）**：
```typescript
// 推广活动
PromotionTier / PromotionCurrentTier / UserPromotion / PromotionStatusResponse

// 邀请返利
ReferralStats / ReferralInfoResponse / ReferralInvite

// 支付/发票
RedeemCodeType（含 'invitation' 值）
// 及所有 Invoice / Payment 相关接口类型
```

**操作建议**：以 zyp-dev 版本为基础，将 v0.1.85 新增的 Sora/Codex 字段（如 `user_agent`、`cache_ttl_overridden`、`WindowStats` 等）追加到末尾，消除重复定义。

---

#### `frontend/src/i18n/locales/en.ts` 和 `zh.ts`

**风险**：双方都新增了大量 i18n key，若取任意一侧将丢失另一侧的文案。

**必须保留（zyp-dev 侧）**：以下命名空间的全量 key：
```
billing.*      (15+ 个导出相关 key)
invoice.*      (50+ 个发票功能 key)
referral.*     (40+ 个邀请返利 key)
promotion.*    (20+ 个推广活动 key)
announcements.*
admin.invoices.*
admin.announcements.*
```

**操作建议**：强制手工合并，以 zyp-dev 为基础文件，将 v0.1.85 的新增 key（`sora.*`、`exportRecords`、`searchUsers` 等）追加进去。**禁止**使用 `accept theirs` 处理 i18n 文件。

---

#### `frontend/src/components/admin/account/AccountTableFilters.vue`

**风险**：v0.1.85 扩展了过滤器（新增 type/group/status 含 rate_limited + Sora/Antigravity）；zyp-dev 当前的过滤器配置（Platform: openai/anthropic/gemini；Status: active/error）是经过裁剪的特定配置。

**必须保留（zyp-dev 侧）**：
- 当前 Platform 下拉：`openai`、`anthropic`、`gemini`（不强制引入 antigravity/sora，评估后再决定是否追加）
- 当前 Status 下拉：`active`、`error`（注意 zyp-dev 不含 `rate_limited`，如需引入须评估 UI 影响）

**操作建议**：先保留 zyp-dev 版本的过滤配置，将 v0.1.85 的扩展过滤项作为**可选追加**处理，不强制引入未在 zyp-dev 中规划的状态值。

---

#### `frontend/src/components/account/EditAccountModal.vue`

**风险**：v0.1.85 新增 OpenAI passthrough 禁用提示；zyp-dev 有模型限制 UI（Whitelist/Mapping 模式）、Custom Error Codes、配额控制（Window Cost、Session 限制、TLS Fingerprint）。

**必须保留（zyp-dev 侧）**：
```
- 模型限制：Whitelist 模式 / Mapping 模式（含预设映射）
- Custom Error Codes 配置
- Quota Control：Window Cost limit + Sticky Reserve
- Session 限制：maxSessions + idleTimeout
- TLS Fingerprint 控制
```

**操作建议**：取 zyp-dev 版本为基础，将 v0.1.85 的 passthrough 禁用提示追加到 OpenAI 透传开关区域，不改动 zyp-dev 已有的配置 UI。

---

### 8.3 合并时的决策优先级规则

| 冲突场景 | 决策规则 |
|---|---|
| v0.1.85 **新增字段/函数/路由**，zyp-dev 未涉及 | 追加到 zyp-dev 版本中 |
| v0.1.85 **修改**了 zyp-dev **也修改**的同一行 | 手工合并，保留双方语义，优先不破坏 zyp-dev 行为 |
| v0.1.85 **删除**了 zyp-dev 中存在的内容 | **必须保留** zyp-dev 版本，拒绝该删除 |
| 生成文件（`wire_gen.go`、`mutation.go`）| 先合并源文件，重新生成，不手工接受任何一侧 |
| i18n 文件（`en.ts`、`zh.ts`） | **禁止** `accept theirs`，只能手工追加 v0.1.85 新增 key |
| 纯 v0.1.85 新文件（Sora 核心文件）| `accept theirs`，不影响已有功能 |

---

### 8.4 功能回归验收（zyp-dev 专项）

合并完成后，必须验证以下 zyp-dev 独有功能路径正常：

```bash
# 1. 后端编译（确保所有 zyp-dev handler/service 编译通过）
go build ./backend/...

# 2. 配置加载（确保 zyp-dev 新增配置字段可正常反序列化）
go test ./backend/internal/config/...

# 3. 路由注册（确保支付/发票/推广/返利/钉钉路由存在）
# 启动后检查：
curl -s http://localhost:PORT/api/v1/admin/payment/orders/summary
curl -s http://localhost:PORT/api/v1/admin/invoices
curl -s http://localhost:PORT/api/v1/user/referral
curl -s http://localhost:PORT/api/v1/user/invoices

# 4. Admin 只读 Key 权限（确保只读 Key 只能访问白名单路径）
# 使用只读 Key 请求非白名单路径，应返回 403

# 5. 前端类型检查（确保 zyp-dev 业务类型不丢失）
pnpm --dir frontend run typecheck

# 6. 前端 i18n 完整性（确保所有业务 key 存在）
pnpm --dir frontend run lint:check
```

**人工核对清单**：
- [ ] 用户侧发票入口可进入，申请流程正常
- [ ] 用户侧邀请返利页面可进入，邀请码显示正常
- [ ] 管理侧发票列表可访问，审批/拒绝/开票操作可用
- [ ] 管理侧公告管理可访问
- [ ] 钉钉机器人 webhook 可接收消息
- [ ] 支付订单列表和导出功能正常
- [ ] 只读 Admin Key 无法访问写接口（返回 403）
- [ ] 账号监控（Anthropic API Key 自动检测）服务正常启动
- [ ] 告警服务（DingTalk 通知）路径无编译报错
