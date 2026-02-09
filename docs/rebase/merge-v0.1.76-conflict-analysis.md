# 合并 v0.1.76 到 zyp-dev 冲突分析报告

**日期**: 2026-02-09
**源分支**: `zyp-dev` (当前分支, 244 独有提交)
**目标 Tag**: `v0.1.76` (360 独有提交)
**冲突文件数**: 61 个
**冲突块总数**: ~130 个

---

## 一、两侧主要变更概述

### zyp-dev 分支独有功能

| 功能模块 | 说明 |
|---------|------|
| 支付系统 | 微信支付、Stripe 接入，在线充值、订单管理 |
| 发票功能 | 开票申请、邮件通知、用户侧发票记录 |
| 钉钉告警 | 账号异常告警、充值通知、钉钉机器人充值 |
| 账号监控 | 自动测试账号可用性、403/503 异常处理、指定账号监控 |
| 429 限流控制 | 变量控制限流时长 |
| 只读 Admin Key | 只读管理接口 |
| 使用记录清理 | 配置化清理策略 |
| UI 主题 | 全局主题色调整 |
| 邀请码(简化版) | `invite_code` 字段（简单文本） |

### v0.1.76 独有功能

| 功能模块 | 说明 |
|---------|------|
| Scope 重构 | 移除 scope 级别概念，替换为 model 级别速率限制 |
| Antigravity 增强 | upstream 账号类型、混合调度、自定义错误策略、线性延迟切换 |
| TOTP 双因素认证 | 用户登录 2FA 支持 |
| 分组排序 | 拖拽排序分组、分组搜索功能 |
| 配额管理 | API Key 配额跟踪 (`IncrementQuotaUsed`) |
| 邀请码(完整版) | `invitation_code` 系统 + `promo_code` 促销码验证 |
| 设置增强 | 首页内容自定义、隐藏 CCS 导入按钮、订阅购买设置 |
| LinuxDo OAuth | 第三方 OAuth 登录 |
| 调度器优化 | 防惊群效应(thundering herd)、粘性会话改进 |
| 客户端断连处理 | streaming 断连后继续消耗上游以确保计费 |

---

## 二、冲突分类汇总

### 冲突类型说明

| 类型 | 含义 |
|------|------|
| **additive-both** | 双方在同一位置各自添加了不同代码 |
| **modify-same** | 双方修改了同一行代码 |
| **delete-vs-add** | 一方无改动(空)，另一方添加了新代码 |

### 建议解决策略说明

| 策略 | 含义 |
|------|------|
| **accept-theirs** | 采用 v0.1.76 版本（上游更完善） |
| **accept-ours** | 保留 zyp-dev 版本（保留定制功能） |
| **manual-merge** | 需手动合并双方代码 |
| **regenerate** | 自动生成文件，合并后重新生成 |

---

## 三、详细冲突分析

### 3.1 文档与配置 (3 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `README.md` | 1 | modify-same | Go 1.25.5, GORM | Go 1.25.7, Ent | **accept-theirs** |
| `README_CN.md` | 1 | modify-same | Go 1.25.5, GORM | Go 1.25.7, Ent | **accept-theirs** |
| `deploy/.env.example` | 1 | additive-both | 添加了大量 zyp-dev 独有配置(钉钉告警、定价、并发、Token刷新等) | 添加了 TOTP 加密密钥配置 | **manual-merge** |

**deploy/.env.example 详细分析**:
- **zyp-dev** 在 JWT 配置后添加了: Default Settings、Rate Limiting、Pricing Data Source、DingTalk Alerts、DingTalk Bot、Concurrency、Token Refresh 等约 112 行配置
- **v0.1.76** 在同一位置添加了 TOTP 加密密钥配置约 11 行
- **解决方案**: 保留两侧全部配置，TOTP 配置放在 JWT 配置之后，zyp-dev 的配置依次排列

---

### 3.2 依赖管理 (2 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/go.mod` | 1 | modify-same | 较低版本依赖 | 较新版本依赖 | **accept-theirs** |
| `backend/go.sum` | 2 | modify-same | golang.org/x 系列 v0.30~v0.39 | golang.org/x 系列 v0.31~v0.40 | **regenerate** |

**go.mod/go.sum 详细分析**:
- 两侧使用了不同版本的 `golang.org/x/*` 系列包
- v0.1.76 版本更新 (`mod v0.31.0 vs v0.30.0`, `net v0.49.0 vs v0.48.0`, `sys v0.40.0 vs v0.39.0` 等)
- **解决方案**: 采用 v0.1.76 的 go.mod，然后运行 `go mod tidy` 重新生成 go.sum

---

### 3.3 Schema/ORM 变更 (7 个文件)

**核心冲突根因**: v0.1.76 将所有 schema 中的硬编码字符串默认值提取到了 `internal/domain` 包中的常量，例如 `"active"` → `domain.StatusActive`，`"anthropic"` → `domain.PlatformAnthropic`。这导致所有 schema 文件在 import 和 Default() 调用处产生冲突。

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/ent/schema/account.go` | 2 | modify-same | `Default("active")` 硬编码 | 导入 `domain` 包，使用 `Default(domain.StatusActive)` | **accept-theirs** |
| `backend/ent/schema/api_key.go` | 2 | modify-same | `Default("active")` 硬编码 | 使用 `Default(domain.StatusActive)` | **accept-theirs** |
| `backend/ent/schema/group.go` | 3 | modify-same | `Default("active")`, `Default("anthropic")`, `Default("standard")` | 使用 `domain.StatusActive`, `domain.PlatformAnthropic`, `domain.SubscriptionTypeStandard` | **accept-theirs** |
| `backend/ent/schema/promo_code.go` | 2 | modify-same | `Default("active")` 硬编码 | 使用 `domain.PromoCodeStatusActive` | **accept-theirs** |
| `backend/ent/schema/redeem_code.go` | 3 | modify-same | `Default("balance")`, `Default("unused")` | 使用 `domain.RedeemTypeBalance`, `domain.StatusUnused` | **accept-theirs** |
| `backend/ent/schema/user.go` | 3 | modify-same | `Default("user")`, `Default("active")` | 使用 `domain.RoleUser`, `domain.StatusActive` | **accept-theirs** |
| `backend/ent/schema/user_subscription.go` | 2 | modify-same | `Default("active")` 硬编码 | 使用 `domain.SubscriptionStatusActive` | **accept-theirs** |

**Schema 冲突详细分析**:
- **所有 7 个 schema 文件的冲突模式完全一致**: 都是 `domain` 包常量提取重构
- 每个文件第一个冲突是 import 语句（v0.1.76 添加了 `"github.com/Wei-Shaw/sub2api/internal/domain"` 导入）
- 后续冲突是字段默认值从字符串字面量改为 domain 常量
- **解决方案**: 全部采用 v0.1.76 版本，确保 `internal/domain` 包存在且包含所有引用的常量
- **注意**: 合并 schema 后需要运行 `go generate ./ent` 重新生成 ent 代码

---

### 3.4 Wire/依赖注入 (4 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/cmd/server/wire_gen.go` | 7 | additive-both | 添加了告警服务、监控服务注入 | 添加了配额服务、TOTP 服务、调度器快照等注入 | **regenerate** |
| `backend/internal/handler/wire.go` | 8 | additive-both | 添加了监控相关 handler 注入 | 添加了 TOTP、配额、错误透传等 handler 注入 | **manual-merge** |
| `backend/internal/repository/wire.go` | 1 | additive-both | 添加了监控 repo 注入 | 添加了配额 repo 注入 | **manual-merge** |
| `backend/internal/service/wire.go` | 1 | additive-both | 添加了告警服务、监控服务注入 | 添加了 TOTP 服务、配额服务等注入 | **manual-merge** |

**Wire 冲突解决策略**:
- `wire_gen.go` 是自动生成文件，手动合并 `wire.go` 后运行 `wire ./cmd/server/` 重新生成
- 核心是合并各层的 `wire.go` 文件中的 Provider 集合

---

### 3.5 Handler 层 (7 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/internal/handler/admin/setting_handler.go` | 1 | delete-vs-add | 无新增 | 添加了 `home_content`, `hide_ccs_import_button`, `purchase_subscription_*` 设置字段 | **accept-theirs** |
| `backend/internal/handler/api_key_handler.go` | 2 | additive-both | 无新增 | 添加了 `QuotaLimit` 字段处理和 `Status` 字段 | **accept-theirs** |
| `backend/internal/handler/auth_handler.go` | 1 | additive-both | 使用 `invite_code` 字段 | 使用 `promo_code` + `invitation_code` 字段 | **manual-merge** |
| `backend/internal/handler/gateway_handler.go` | 7 | additive-both | 添加了监控测试接口逻辑 | 添加了 upstream 转发、错误策略处理、混合调度等 | **manual-merge** |
| `backend/internal/handler/gemini_v1beta_handler.go` | 1 | delete-vs-add | 无新增 | 添加了 `antiGravityGatewayService` 字段用于 Antigravity 网关 | **accept-theirs** |
| `backend/internal/handler/handler.go` | 3 | additive-both | 添加了监控 handler 字段 | 添加了 TOTP、配额管理 handler 字段 | **manual-merge** |
| `backend/cmd/jwtgen/main.go` | 1 | modify-same | 使用 `scope` 声明 | 移除了 scope 使用 | **accept-theirs** |

**Handler 冲突根因分析**:
- `gateway_handler.go` 是最复杂的冲突文件之一（7 个冲突块），具体冲突点：
  1. **请求解析**(L226): v0.1.76 添加了 `ParseGatewayRequest`、Claude Code 客户端检测、Thinking 状态上下文
  2. **拦截处理**(L371): v0.1.76 用 `detectInterceptType` 替代了简单的 warmup 检查
  3. **Forward 调用签名**(L452): v0.1.76 使用 `requestCtx` 传递 account switch count
  4. **使用量记录**(L491): v0.1.76 改为异步 goroutine 记录，使用独立的 background context
  5. **主请求循环**(L531): v0.1.76 重构了错误处理（`lastFailoverErr` 类型化错误 + `forceCacheBilling`）
  6. **重复记录**(L819): v0.1.76 删除了同步记录（已由异步处理）
  7. **OpenAI 网关**(L1256): v0.1.76 添加了相同的 Claude Code 客户端检测逻辑
  - **建议**: 以 v0.1.76 为基础，将 zyp-dev 的监控测试逻辑补充进去
- `jwtgen/main.go` - v0.1.76 修改了 `NewAuthService` 构造函数签名（参数数量从 8 → 9），需与最终合并后的签名保持一致

---

### 3.6 Service 层 (13 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/internal/service/admin_service.go` | 5 | additive-both | 添加了监控、告警相关逻辑 | 添加了分组搜索、错误透传规则、模型配额等 | **manual-merge** |
| `backend/internal/service/admin_service_group_test.go` | 1 | delete-vs-add | 无新增 | 添加了 `TestAdminService_ListGroups_WithSearch` 测试 | **accept-theirs** |
| `backend/internal/service/api_key.go` | 1 | delete-vs-add | 无新增 | 添加了 `ListKeysByUserID`, `ListKeysByGroupID`, `IncrementQuotaUsed` 接口方法 | **accept-theirs** |
| `backend/internal/service/api_key_service_delete_test.go` | 1 | delete-vs-add | 无新增 | 添加了测试桩的新接口实现 (`ListKeysByUserID`, `ListKeysByGroupID`, `IncrementQuotaUsed`) | **accept-theirs** |
| `backend/internal/service/auth_service.go` | 3 | additive-both | 使用 `InviteCode` 字段 | 使用 `PromoCode` + `InvitationCode` + TOTP 验证 | **manual-merge** |
| `backend/internal/service/gateway_service.go` | 4 | additive-both | 添加了监控测试、告警逻辑 | 添加了 upstream 转发、错误策略处理 | **manual-merge** |
| `backend/internal/service/gemini_messages_compat_service.go` | 1 | delete-vs-add | 无新增 | 添加了大量方法: `resolvePlatformAndSchedulingMode`, `tryStickySessionHit`, `isAccountUsableForRequest`, `selectBestGeminiAccount`, `isBetterGeminiAccount` 等 (~250 行) | **accept-theirs** |
| `backend/internal/service/group.go` | 2 | additive-both | 添加了 `ModelRoutingEnabled` 字段 | 添加了 `FallbackGroupIDOnInvalidRequest`, `SupportedModelScopes`, `SortOrder`, `MCPXmlInject`, `ClaudeCodeOnly`, `CopyAccountsFromGroupIDs` 等字段 | **manual-merge** |
| `backend/internal/service/ratelimit_service.go` | 2 | modify-same | 使用 scope 级别限流 | 重构为 model 级别限流 (`PreCheckUsage`) | **accept-theirs** |
| `backend/internal/service/setting_service.go` | 1 | delete-vs-add | 无新增 | 添加了 `HomeContent`, `HideCCSImportButton`, `PurchaseSubscription*` 设置字段 | **accept-theirs** |
| `backend/internal/service/token_refresh_service.go` | 2 | additive-both | 添加了 `accountAlert` (告警服务) | 添加了 `schedulerCache` (调度器缓存同步) | **manual-merge** |

**Service 层冲突根因分析**:
- `ratelimit_service.go` - v0.1.76 进行了 **scope→model 级别重构**，这是一个破坏性变更，必须采用 v0.1.76 版本
  - 冲突 1 (handle429): HEAD 有自定义 fallback cooldown + Sonnet model-scoped 限流；v0.1.76 有多平台处理(OpenAI x-codex/Gemini body 解析)
  - 冲突 2 (parse failure): HEAD 用可配置 fallbackResetAt；v0.1.76 用固定 5 分钟
  - **建议**: 以 v0.1.76 为基础，可选保留 HEAD 的可配置 fallback cooldown
- `gateway_service.go` - 4 个冲突都在粘性会话(sticky session)逻辑：v0.1.76 引入了 `shouldClearStickySession` + `IsSchedulableForModelWithContext` + 错误透传规则匹配等更精细的调度
- `auth_service.go` - 3 个冲突，核心差异：
  - 注册函数签名: HEAD `(email, password, verifyCode, inviteCode)` vs v0.1.76 `(email, password, verifyCode, promoCode, invitationCode)`
  - 验证流程: v0.1.76 增加了邀请码验证(基于 redeem) + 服务可用性检查
  - 注册后操作: HEAD 有 promotion 初始化 + referral 记录；v0.1.76 有 invitation code 标记 + promo code 应用
  - **建议**: 合并时需保留两侧的后处理逻辑
- `admin_service.go` - 5 个冲突，主要在 AdminService 结构体字段和构造函数参数差异：HEAD 有 `balanceService/accountAlert/paymentOrderRepo/cfg`；v0.1.76 有 `userGroupRateRepo/proxyProber`
- `gemini_messages_compat_service.go` - v0.1.76 添加了 ~250 行的调度逻辑重构，提取了多个清晰的辅助方法，直接采用即可

---

### 3.7 Repository 层 (3 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/internal/repository/claude_usage_service.go` | 1 | modify-same | 内联实现 `FetchUsage`，直接使用 `httpclient.GetClient` | 重构为委托调用 `FetchUsageWithOptions`（DRY 优化） | **accept-theirs** |
| `backend/internal/repository/usage_log_repo.go` | 1 | delete-vs-add | 无新增 | 添加了 `imageSize` 和 `reasoningEffort` 可空字段扫描 | **accept-theirs** |
| `backend/internal/repository/wire.go` | 1 | additive-both | 添加了监控 repo | 添加了配额 repo | **manual-merge** |

---

### 3.8 路由与中间件 (4 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/internal/server/routes/admin.go` | 2 | additive-both | 添加了监控相关路由 | 添加了分组排序、TOTP、错误透传、配额等路由 | **manual-merge** |
| `backend/internal/server/routes/user.go` | 1 | delete-vs-add | 无新增 | 添加了 TOTP 设置路由、Key 配额路由 | **accept-theirs** |
| `backend/internal/server/api_contract_test.go` | 2 | delete-vs-add | 无新增 | 添加了新路由的契约测试 | **accept-theirs** |
| `backend/internal/server/middleware/api_key_auth_test.go` | 1 | delete-vs-add | 无新增 | 添加了 Key 状态检查测试 | **accept-theirs** |
| `backend/internal/server/middleware/api_key_auth_google_test.go` | 1 | delete-vs-add | 无新增 | 添加了 Google Key 状态检查测试 | **accept-theirs** |

---

### 3.9 Pkg 层 (2 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `backend/internal/pkg/antigravity/request_transformer.go` | 1 | additive-both | 导入了 `"os"` | 导入了 `"math/rand"` + `"strconv"` | **manual-merge** |
| `backend/internal/pkg/claude/constants.go` | 1 | additive-both | 添加了 `DefaultMonitorModel` 常量（账号监控用） | 添加了 `ModelIDOverrides`/`ModelIDReverseOverrides` map 和 `NormalizeModelID`/`DenormalizeModelID` 函数（Claude OAuth 模型 ID 标准化） | **manual-merge** |

**Pkg 层冲突分析**: 两个文件的冲突都是双方独立新增代码，互不影响，保留双方即可。

---

### 3.10 前端 API 层 (3 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `frontend/src/api/admin/index.ts` | 3 | additive-both | 添加了监控 API | 添加了分组排序、错误透传、配额管理 API | **manual-merge** |
| `frontend/src/api/admin/settings.ts` | 1 | delete-vs-add | 无新增 | 添加了 `home_content`, `hide_ccs_import_button`, `purchase_subscription_*` 设置 | **accept-theirs** |
| `frontend/src/api/index.ts` | 1 | additive-both | 使用 `invite_code` | 使用 `promo_code` + `invitation_code` | **manual-merge** |

---

### 3.11 前端组件 (5 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `frontend/src/components/account/AccountStatusIndicator.vue` | 2 | additive-both | 旧版状态显示 | 新增了更多状态类型和颜色映射 | **accept-theirs** |
| `frontend/src/components/account/CreateAccountModal.vue` | 1 | modify-same | Antigravity 仅 OAuth 类型(固定显示) | Antigravity 支持 OAuth + Upstream 类型(可切换) | **accept-theirs** |
| `frontend/src/components/admin/usage/UsageTable.vue` | 3 | additive-both | 旧版用量表格 | 新增模型级别用量展示、配额列等 | **accept-theirs** |
| `frontend/src/components/layout/AppSidebar.vue` | 2 | additive-both | 旧版侧边栏 | 新增了订阅购买入口、布局调整 | **accept-theirs** |
| `frontend/src/composables/useModelWhitelist.ts` | 1 | modify-same | 旧版白名单 | 更新了模型白名单列表 | **accept-theirs** |

---

### 3.12 前端视图 (9 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `frontend/src/views/admin/AccountsView.vue` | 1 | modify-same | 使用 `formatDate`, `Group` 类型 | 使用 `formatDateTime`, `AdminGroup` 类型, 新增 `ErrorPassthroughRulesModal` | **accept-theirs** |
| `frontend/src/views/admin/GroupsView.vue` | 2 | additive-both | 添加了 `ModelRoutingEnabled` | 添加了降级分组选项 computed、`fallback_group_id_on_invalid_request`, `supported_model_scopes`, `mcp_xml_inject`, `copy_accounts_from_group_ids` | **manual-merge** |
| `frontend/src/views/admin/SettingsView.vue` | 1 | delete-vs-add | 无新增 | 添加了 `home_content`, `hide_ccs_import_button`, `purchase_subscription_*` 设置提交 | **accept-theirs** |
| `frontend/src/views/auth/EmailVerifyView.vue` | 1 | additive-both | 使用 `invite_code` | 使用 `promo_code` + `invitation_code` | **manual-merge** |
| `frontend/src/views/auth/LoginView.vue` | 1 | additive-both | 引入 `Modal` 组件 | 引入 `LinuxDoOAuthSection` + `TotpLoginModal` | **manual-merge** |
| `frontend/src/views/auth/RegisterView.vue` | 6 | additive-both | 使用简单 `invite_code` 字段 | 使用完整的 `invitation_code` 系统 + `promo_code` 验证 + debounce + `onUnmounted` 清理 | **manual-merge** |
| `frontend/src/views/user/KeysView.vue` | 2 | additive-both | 简化的状态 badge | 更丰富的状态 badge(含过期等) + CCS 导入按钮 | **accept-theirs** |
| `frontend/vite.config.ts` | 1 | modify-same | 固定端口 3000，server 在外层 | 动态 `devPort`，server 在内层 | **accept-theirs** |

---

### 3.13 前端类型/路由/i18n (5 个文件)

| 文件 | 冲突数 | 冲突类型 | zyp-dev | v0.1.76 | 建议策略 |
|------|--------|---------|---------|---------|---------|
| `frontend/src/types/index.ts` | 1 | additive-both | 添加了监控相关类型 | 添加了配额、TOTP、邀请码等类型 | **manual-merge** |
| `frontend/src/router/index.ts` | 1 | delete-vs-add | 无新增 | 添加了 TOTP 相关路由 | **accept-theirs** |
| `frontend/src/i18n/locales/en.ts` | 2 | additive-both | 添加了监控、告警相关翻译 | 添加了 TOTP、配额、邀请码等翻译 | **manual-merge** |
| `frontend/src/i18n/locales/zh.ts` | 3 | additive-both | 添加了监控、告警相关中文翻译 | 添加了 TOTP、配额、邀请码等中文翻译 | **manual-merge** |

---

## 四、核心冲突根因分析

### 4.1 Scope → Model 级别重构 (高优先级)

v0.1.76 进行了一次重要的架构重构：**移除 scope 级别的速率限制，替换为 model 级别**。这影响了：
- `ratelimit_service.go` - 核心限流逻辑重写
- `claude_usage_service.go` - 使用量统计逻辑变更
- `jwtgen/main.go` - JWT 生成中 scope 声明移除
- 多个测试文件

**影响**: 这是一个破坏性的架构变更，必须采用 v0.1.76 的实现。zyp-dev 的旧 scope 逻辑与新架构不兼容。

### 4.2 邀请码系统差异 (中优先级)

- **zyp-dev**: 使用简单的 `invite_code` 文本字段
- **v0.1.76**: 使用完整的 `invitation_code` 系统 + `promo_code` 促销码验证系统

**影响**: 需要决定是否保留 zyp-dev 的简单邀请码还是迁移到 v0.1.76 的完整系统。建议采用 v0.1.76 的系统，然后将 zyp-dev 的 `invite_code` 作为 `invitation_code` 的别名或迁移数据。

### 4.3 zyp-dev 独有功能保留 (需谨慎合并)

以下 zyp-dev 功能在 v0.1.76 中不存在，需要确保合并后保留：
- 钉钉告警服务 (`AccountAlertService`)
- 账号自动监控测试
- 支付系统 (这些文件大多不冲突)
- 发票功能 (这些文件大多不冲突)
- 429 限流时长变量控制

---

## 五、建议合并步骤

### 第一步：处理自动生成文件
```bash
# 先处理非自动生成的冲突文件，最后处理这些：
# - backend/cmd/server/wire_gen.go  →  合并 wire.go 后运行 wire 重新生成
# - backend/go.sum                  →  合并 go.mod 后运行 go mod tidy
```

### 第二步：批量 accept-theirs (约 25 个文件)
以下文件直接采用 v0.1.76 版本：
```bash
git checkout --theirs \
  README.md \
  README_CN.md \
  backend/cmd/jwtgen/main.go \
  backend/ent/schema/api_key.go \
  backend/ent/schema/promo_code.go \
  backend/ent/schema/redeem_code.go \
  backend/ent/schema/user.go \
  backend/ent/schema/user_subscription.go \
  backend/go.mod \
  backend/internal/handler/admin/setting_handler.go \
  backend/internal/handler/api_key_handler.go \
  backend/internal/handler/gemini_v1beta_handler.go \
  backend/internal/pkg/antigravity/request_transformer.go \
  backend/internal/pkg/claude/constants.go \
  backend/internal/repository/claude_usage_service.go \
  backend/internal/repository/usage_log_repo.go \
  backend/internal/server/api_contract_test.go \
  backend/internal/server/middleware/api_key_auth_google_test.go \
  backend/internal/server/middleware/api_key_auth_test.go \
  backend/internal/server/routes/user.go \
  backend/internal/service/admin_service_group_test.go \
  backend/internal/service/api_key.go \
  backend/internal/service/api_key_service_delete_test.go \
  backend/internal/service/gemini_messages_compat_service.go \
  backend/internal/service/ratelimit_service.go \
  backend/internal/service/setting_service.go \
  backend/internal/service/api_key_service_delete_test.go \
  frontend/src/api/admin/settings.ts \
  frontend/src/components/account/AccountStatusIndicator.vue \
  frontend/src/components/account/CreateAccountModal.vue \
  frontend/src/components/admin/usage/UsageTable.vue \
  frontend/src/components/layout/AppSidebar.vue \
  frontend/src/composables/useModelWhitelist.ts \
  frontend/src/router/index.ts \
  frontend/src/views/admin/AccountsView.vue \
  frontend/src/views/admin/SettingsView.vue \
  frontend/src/views/user/KeysView.vue \
  frontend/vite.config.ts

git add <上述文件>
```

### 第三步：手动合并 (约 25 个文件)
以下文件需要手动合并，保留双方功能：

**高优先级（核心业务逻辑）**:
1. `backend/internal/handler/gateway_handler.go` - 合并监控测试 + upstream 转发
2. `backend/internal/service/gateway_service.go` - 合并监控 + upstream 逻辑
3. `backend/internal/service/admin_service.go` - 合并监控 + 分组搜索
4. `backend/internal/service/auth_service.go` - 合并邀请码系统
5. `backend/internal/handler/auth_handler.go` - 合并注册参数处理

**中优先级（DI/路由/Schema）**:
6. `backend/ent/schema/account.go` - 保留双方新增字段
7. `backend/ent/schema/group.go` - 保留双方新增字段
8. `backend/internal/handler/handler.go` - 合并 handler 字段
9. `backend/internal/handler/wire.go` - 合并 Provider
10. `backend/internal/repository/wire.go` - 合并 Provider
11. `backend/internal/service/wire.go` - 合并 Provider
12. `backend/internal/server/routes/admin.go` - 合并路由
13. `backend/internal/service/token_refresh_service.go` - 保留 `accountAlert` + `schedulerCache`
14. `backend/internal/service/group.go` - 保留双方新增字段

**低优先级（前端/配置）**:
15. `deploy/.env.example` - 合并配置
16. `frontend/src/api/admin/index.ts` - 合并 API
17. `frontend/src/api/index.ts` - 合并注册参数
18. `frontend/src/types/index.ts` - 合并类型定义
19. `frontend/src/i18n/locales/en.ts` - 合并翻译
20. `frontend/src/i18n/locales/zh.ts` - 合并翻译
21. `frontend/src/views/admin/GroupsView.vue` - 合并分组表单字段
22. `frontend/src/views/auth/RegisterView.vue` - 合并注册表单
23. `frontend/src/views/auth/LoginView.vue` - 合并登录组件引入
24. `frontend/src/views/auth/EmailVerifyView.vue` - 合并验证参数

### 第四步：重新生成自动文件
```bash
cd backend
go mod tidy                    # 重新生成 go.sum
go generate ./ent              # 重新生成 ent 代码
wire ./cmd/server/             # 重新生成 wire_gen.go
```

### 第五步：编译验证
```bash
cd backend && go build ./...   # 编译检查
cd frontend && pnpm install && pnpm build  # 前端构建检查
```

---

## 六、风险评估

| 风险项 | 级别 | 说明 |
|-------|------|------|
| Scope→Model 重构不彻底 | **高** | 确保所有 scope 相关代码已迁移到 model 级别 |
| Wire DI 注入遗漏 | **高** | 合并 wire.go 后需确保所有新服务都被正确注入 |
| Schema 迁移 | **高** | 新增字段需要数据库迁移脚本 |
| 邀请码数据迁移 | **中** | 如果有用户已使用旧 invite_code，需要数据迁移 |
| 前端类型不一致 | **中** | `Group` vs `AdminGroup` 等类型变更 |
| 翻译缺失 | **低** | 合并后检查 i18n 键是否完整 |

---

## 七、总结

本次合并涉及两个长期分叉的分支，zyp-dev 专注于**运营功能**(支付/发票/告警/监控)，v0.1.76 专注于**核心功能增强**(scope重构/antigravity/TOTP/配额)。

- **约 36 个文件**可直接采用 v0.1.76 版本 (accept-theirs)
- **约 25 个文件**需手动合并
- 合并后需重新生成 wire_gen.go, go.sum, ent 代码
- 建议在合并后进行全面的编译和功能测试
