# 合并报告：`v0.1.103` 到 `zyp-dev`

## 1. 合并背景

- 合并日期：2026-03-19（UTC）
- 目标分支：`zyp-dev`
- 来源标签：`v0.1.103`（`3cedfcd827809cb9d76196ee67dcc3480b5100b7`）
- 合并前目标分支 HEAD：`6facef2858e477e9017e24c2080ee792ea2f88ed`
- 共同祖先（merge-base）：`7be5e1734c1c0882c851f9fd1c2ee70cc62b7d29`
- 分叉规模（`git rev-list --left-right --count HEAD...v0.1.103`）：
- 当前分支独有提交：`287`
- Tag 独有提交：`649`

## 2. 安全措施与冲突处理策略

- 合并前已创建备份分支：`backup/pre-merge-v0.1.103-20260319`
- 实际合并命令：`git merge --no-commit --no-ff v0.1.103`
- 处理原则：**确保当前分支已有修改不被覆盖**
- 冲突解决方式：
- 所有 `UU`（双方都修改）：统一采用 `ours`，保留 `zyp-dev` 版本
- `DU`（当前分支删除、tag 修改）：保留当前分支删除结果

## 3. 冲突点清单

冲突总数：`55`

- `UU`（双方都修改）：`54`
- `DU`（当前分支删除、tag 修改）：`1`

### 3.1 基础设施 / 构建 / 部署类冲突

- `UU` `.gitignore`
- `UU` `backend/cmd/jwtgen/main.go`
- `UU` `backend/cmd/server/wire.go`
- `UU` `backend/cmd/server/wire_gen.go`
- `UU` `backend/go.sum`
- `UU` `deploy/config.example.yaml`
- `UU` `deploy/docker-compose.yml`
- `DU` `tools/check_pnpm_audit_exceptions.py`

### 3.2 后端业务 / 路由 / 服务层冲突

- `UU` `backend/internal/config/config.go`
- `UU` `backend/internal/handler/admin/setting_handler.go`
- `UU` `backend/internal/handler/auth_handler.go`
- `UU` `backend/internal/handler/dto/settings.go`
- `UU` `backend/internal/handler/openai_gateway_handler.go`
- `UU` `backend/internal/handler/setting_handler.go`
- `UU` `backend/internal/repository/account_repo.go`
- `UU` `backend/internal/repository/usage_log_repo.go`
- `UU` `backend/internal/server/middleware/admin_auth_test.go`
- `UU` `backend/internal/server/routes/admin.go`
- `UU` `backend/internal/server/routes/gateway.go`
- `UU` `backend/internal/service/account_test_service.go`
- `UU` `backend/internal/service/account_usage_service.go`
- `UU` `backend/internal/service/admin_service.go`
- `UU` `backend/internal/service/auth_service.go`
- `UU` `backend/internal/service/billing_service.go`
- `UU` `backend/internal/service/dashboard_service.go`
- `UU` `backend/internal/service/domain_constants.go`
- `UU` `backend/internal/service/ratelimit_service.go`
- `UU` `backend/internal/service/setting_service.go`
- `UU` `backend/internal/service/token_refresh_service.go`
- `UU` `backend/internal/service/wire.go`
- `UU` `backend/internal/web/embed_on.go`
- `UU` `backend/resources/model-pricing/model_prices_and_context_window.json`

### 3.3 前端 UI / 状态管理 / 多语言冲突

- `UU` `frontend/src/App.vue`
- `UU` `frontend/src/api/admin/dashboard.ts`
- `UU` `frontend/src/api/admin/settings.ts`
- `UU` `frontend/src/components/account/AccountUsageCell.vue`
- `UU` `frontend/src/components/account/BulkEditAccountModal.vue`
- `UU` `frontend/src/components/account/EditAccountModal.vue`
- `UU` `frontend/src/components/admin/account/AccountBulkActionsBar.vue`
- `UU` `frontend/src/components/admin/account/AccountTableFilters.vue`
- `UU` `frontend/src/components/common/AnnouncementBell.vue`
- `UU` `frontend/src/components/layout/AppHeader.vue`
- `UU` `frontend/src/components/layout/AppSidebar.vue`
- `UU` `frontend/src/i18n/locales/en.ts`
- `UU` `frontend/src/i18n/locales/zh.ts`
- `UU` `frontend/src/main.ts`
- `UU` `frontend/src/stores/app.ts`
- `UU` `frontend/src/types/index.ts`
- `UU` `frontend/src/views/admin/AccountsView.vue`
- `UU` `frontend/src/views/admin/DashboardView.vue`
- `UU` `frontend/src/views/admin/GroupsView.vue`
- `UU` `frontend/src/views/admin/SettingsView.vue`
- `UU` `frontend/src/views/admin/UsageView.vue`
- `UU` `frontend/src/views/auth/RegisterView.vue`
- `UU` `frontend/src/views/user/UsageView.vue`

## 4. 冲突原因分析

- 分支长期并行开发，和共同祖先相比双边改动都很大。
- 改动集中在同一批热点文件（`service`、`route`、`dashboard`、`usage`、`settings`），重叠概率高。
- 依赖锁文件和部署配置（如 `go.sum`、`docker-compose`、配置模板）本身冲突概率就高。
- 前端入口、布局、状态管理、i18n 与管理台页面均被双方同时修改。
- 生命周期差异导致 `DU` 冲突：当前分支已删除某工具脚本，而 tag 侧仍在继续修改该文件。

## 5. 建议的后续处理方案

当前合并策略已保证“冲突文件不覆盖当前分支逻辑”。如需吸收 `v0.1.103` 中有价值的改动，建议做二次选择性回补：

1. 按主题回补（推荐）
- 主题 A：用量统计与上游模型追踪能力
- 主题 B：管理端设置与仪表盘改动
- 主题 C：部署与运行时配置调整
- 做法：按主题比对 tag 差异，仅挑选已验证的 hunk/cherry-pick 回补到 `zyp-dev`

2. 优先复核高风险文件
- Backend：`backend/internal/repository/usage_log_repo.go`
- Backend：`backend/internal/service/dashboard_service.go`
- Backend：`backend/internal/service/setting_service.go`
- Backend：`backend/internal/service/ratelimit_service.go`
- Frontend：`frontend/src/views/admin/UsageView.vue`
- Frontend：`frontend/src/views/admin/DashboardView.vue`
- Frontend：`frontend/src/api/admin/dashboard.ts`
- Frontend：`frontend/src/api/admin/settings.ts`
- Deploy：`deploy/docker-compose.yml`
- Deploy：`deploy/config.example.yaml`

3. 合并后验证清单
- 执行后端单测与集成测试
- 执行前端单测与构建
- 覆盖 admin 的 `dashboard/usage/settings` 冒烟验证
- 验证配置文件与 docker compose 启动流程

## 6. 后续可用命令（可选）

- 对比某个冲突文件与 tag 版本差异：
- `git diff HEAD v0.1.103 -- <file>`
- 临时取 tag 版本文件做人工三方整合：
- `git checkout v0.1.103 -- <file>`
- 查看最近 merge 记录：
- `git log --oneline --merges -n 5`

## 7. 账号自动检测冲突专题（本次补充）

### 7.1 `v0.1.103` 的账号自动检测逻辑

- 采用“**定时自动检测（Scheduled Test）**”链路：
- 后端按计划（plan）周期性执行账号连通性测试（`ScheduledTestRunnerService`）
- 测试结果落库（plan/result）
- 可通过账号恢复接口统一清理 runtime 异常状态（rate limit、overload、temp unsched、error）

### 7.2 当前分支原有逻辑

- 存在“**监控器（monitor）自动检测**”链路：
- `AnthropicAPIKeyMonitorService`
- `OpenAIAPIKeyMonitorService`
- 监控器依据连续成功/失败阈值自动切换可调度状态。

### 7.3 是否存在冲突

- 结论：**存在潜在冲突**（双链路同时写账号 runtime 状态）。
- 冲突表现：
- 同一账号可能被 monitor 与 scheduled runner 同时判定并改写状态
- 状态抖动风险上升（重复恢复、重复暂停、窗口期不一致）

### 7.4 本次已执行的合并处理

- 已将当前分支 monitor 逻辑在 DI 层临时禁用（provider 返回 `nil`）：
- `backend/internal/service/wire.go`
- `ProvideAnthropicAPIKeyMonitorService`
- `ProvideOpenAIAPIKeyMonitorService`
- 已补齐 `v0.1.103` 定时自动检测主链路：
- 定时计划路由与处理器接入（admin 路由）
- 定时 runner 注入与启动（wire provider + cleanup stop）
- 账号恢复/配额重置接口接入
- 前端账号页补齐定时检测面板与动作菜单接线（schedule/recover/reset-quota）

### 7.5 “是否会完全覆盖当前分支已有账号检测逻辑”结论

- 从机制上：`v0.1.103` 的 scheduled test **不会天然“自动覆盖”** monitor 逻辑；两者可并存并冲突。
- 从本次落地结果：已**显式禁用 monitor**，因此运行时由 scheduled test 链路主导，避免双写冲突。

## 8. 本轮按文档落地结果（2026-03-19 UTC）

### 8.1 已完成合并补齐项

- 已补齐 `AccountRepository` 接口缺失实现：
- `ListSchedulableUngroupedByPlatform`
- `ListSchedulableUngroupedByPlatforms`
- 已对齐 dashboard 查询链路签名与能力：
- `GetUsageTrendWithFilters` / `GetModelStatsWithFilters` 增加 `request_type` 参数
- 增加 `GetModelStatsWithFiltersBySource`（支持 `requested/upstream/mapping`）
- 增加 `GetGroupStatsWithFilters`
- 增加 `GetGroupUsageSummary`（调用 `GetAllGroupUsageSummary`）
- 已补齐 group 输入字段并落库映射：
- `SoraStorageQuotaBytes`
- `AllowMessagesDispatch`
- `DefaultMappedModel`
- 已补齐 AdminService 分组倍率能力：
- `GetGroupRateMultipliers`
- `ClearGroupRateMultipliers`
- `BatchSetGroupRateMultipliers`
- 已补齐用户输入字段并接入更新：
- `CreateUserInput/UpdateUserInput` 增加 `SoraStorageQuotaBytes`
- 已补齐 LinuxDo OAuth 邀请码注册链路所需能力：
- `ErrOAuthInvitationRequired`
- `LoginOrRegisterOAuthWithTokenPair(ctx, email, username, invitationCode)`
- `CreatePendingOAuthToken`
- `VerifyPendingOAuthToken`
- 已补回 OpenAI Chat Completions 依赖的 handler 通用方法：
- `recoverResponsesPanic`
- `ensureResponsesDependencies` / `missingResponsesDependencies`
- `acquireResponsesUserSlot`
- `acquireResponsesAccountSlot`
- `ensureOpenAIPoolModeSessionHash`
- 已修正 gateway 路由注册签名漂移：
- `registerRoutes` 调用 `RegisterGatewayRoutes` 参数对齐

### 8.2 编译验证

- 已执行：
- `cd backend && GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-modcache GOPATH=/tmp/go go build ./cmd/server`
- 结果：**通过（exit code 0）**
- 已执行：
- `pnpm -C frontend build`
- 结果：**通过（exit code 0）**

### 8.3 测试回归

- 已执行：
- `cd backend && go test -count=1 ./...`
- 结果：**通过（exit code 0）**
- 说明：该命令在受限沙箱中会因 `httptest` 监听本地端口被拒绝；切换到非沙箱环境后已完整通过。

### 8.4 当前结论

- 文档中“按主题回补并完成构建/测试验证”的合并项已落地。
- 当前状态为：
- 后端全量单测通过
- 前端构建通过
- 用户未跟踪文件（`deploy/docker-compose-db.yaml`、`skill/sub2api.config`）未被修改/覆盖。
