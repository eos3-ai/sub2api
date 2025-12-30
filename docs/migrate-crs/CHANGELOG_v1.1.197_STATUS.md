# CRS v1.1.197 → Sub2API 功能迁移状态

为方便跟踪 `lei-claude-relay-service` `docs/sub2api/CHANGELOG_v1.1.197_to_HEAD.md` 中罗列的 13 项改造在 Sub2API 中的落地情况，下面列出了当前仓库的对照状态及后续差距。

## 快速概览
| # | 功能 | 状态 | 说明 |
|---|------|------|------|
| 1 | 💰 支付系统 | 部分完成 | 已补齐用户端基础 API + `/payment` 前端页面（固定套餐 + 自定义金额、支付宝/微信选择、金额计算弹窗、二维码展示 + 状态轮询）；后端已接入 ZPay/Stripe（可返回 `pay_url/qr_url`）；仍缺独立的支付结果页与回调校验补强。 |
| 2 | 🎁 活动优惠 | 部分完成 | 服务、存储与配置已加入，但无 API/UI 展示。 |
| 3 | 👥 邀请返利 | 部分完成 | 数据模型完成，前台入口与邀请码记录流程缺失。 |
| 4 | 💵 用户余额 | 部分完成 | 余额/流水表与 BalanceService 已有；当前前端以“我的订单”作为充值记录展示；管理员后台加/减/设置余额已接入 BalanceService 写入账本，并在“加余额”时额外生成“后台充值”订单对用户可见；仍缺账本流水的后台查询/导出 UI。 |
| 5 | 📧 邮件服务 | 部分完成 | SMTP/验证码已实现，密码找回与文档缺失。 |
| 6 | 🔐 用户认证增强 | 部分完成 | 注册/登录/改密完成，重置密码/旧格式迁移缺失。 |
| 7 | 🔑 API Key 管理增强 | 未迁移 | 明文查看、加密存储、分钟级统计与审计均未落地。 |
| 8 | 📊 用户管理优化 | 部分完成 | 列表检索可用，已补齐筛选导出；缓存/并发优化等能力未见。 |
| 9 | 🛡️ 安全修复 | 不适用 | Sub2API 采用 JWT/Admin API Key，不存在 Redis 会话缺陷。 |
|10 | 🖥️ 前端界面改造 | 部分完成 | 已按 `docs/migrate-crs/zhifu.png`、`docs/migrate-crs/tanchuang.png` 补齐 `/payment` 样式与交互，并支持二维码展示/支付状态轮询；仍缺独立支付结果页与更完整的支付闭环提示。 |
|11 | 🤖 钉钉机器人 | 未迁移 | 仓库内无任何 dingtalk 相关代码。 |
|12 | 🏗️ 架构调整 | 未迁移（结构不同） | 仅有 `deploy/docker-compose.yml`，未引入 CRS 的多 Compose/脚本。 |
|13 | 📝 其他改进 | 未迁移 | 脚本/文档/工具均缺失。 |

> 状态含义：**已完成**（无需动作）、**部分完成**（已有部分实现但缺能力差距）、**未迁移**（完全缺失）、**不适用**（架构差异无需迁移）。

## 建议执行顺序（面向现有环境）
1. **支付前台闭环（优先出前端效果）**：`/payment` 页面 + 套餐列表/自定义金额/下单/支付二维码展示/状态轮询已具备；下一步优先补齐独立支付结果页、回调安全校验（金额/币种/状态）与失败/过期分支，让链路真正闭环。
2. **余额与流水可视化**：当前前端以“在线充值 → 我的订单”作为充值记录展示；管理员后台加/减/设置余额已写入 `recharge_records`（并在加余额时生成“后台充值”订单对用户可见）；下一步若要对账/审计，补齐 `recharge_records` 的管理员查询/导出 API/UI。
3. **活动优惠展示**：在现有服务上增加用户/管理员查询接口，并在 Dashboard 中放置 `PromotionBanner`、倒计时等组件，把优惠信息可视化。
4. **邀请返利入口**：为注册流程添加邀请码字段，接入 `ReferralService`，补 `/api/v1/referral/*` 接口及前端邀请页，释放既有返利能力。
5. **忘记密码链路**：依托 `EmailService` 增加 `forgot/reset password` API、Redis token，以及对应前端页面，完善基础认证体验。
6. **API Key 明文查看与审计**：实现 AES-256-GCM 存储、Reveal handler、管理员理由校验及审计记录，并提供 `RevealApiKeyModal`，满足安全合规要求。
7. **用户管理优化**：补缓存、并发控制与导出能力，更新 `UsersView`/仪表盘以匹配 changelog 列出的 UI/UX 改进。
8. **钉钉机器人通知**：新增 Dingtalk 配置、handler 与前端配置页面，满足异动提醒需求。
9. **架构与部署资产**：按需补齐多套 Compose、脚本与 Nginx 示例，或在文档中说明替代方案，方便不同环境部署。
10. **脚本与文档收拢**：迁移 changelog 中其余脚本、API 参考及支付说明，完善使用与排障资料。

> 以上顺序遵循 “先保障可直接呈现的新前端体验，再补链路闭环、最后处理基础设施/文档” 的策略，可在当前 Go + Vue 架构下快速落地。

## 下一步实施计划（建议）
### P0：真实支付闭环（后端）
1. **已完成：ZPay/Stripe 基础下单**：可生成 `pay_url/qr_url`（ZPay 收银台链接 / Stripe 微信支付指引页与二维码）。
2. **已完成：回调路由与验签入口**：已提供 ZPay notify 与 Stripe webhook 回调入口，并在支付成功事件中调用 `PaymentService.MarkOrderPaid`。
3. **已完成：订单过期任务**：`payment.enabled=true` 时启动后台定时清理过期订单。
4. **已补强：安全校验（第一阶段）**：回调中已增加金额/币种/订单状态的基础校验，并补齐 Stripe `payment_failed/canceled` 事件将订单落库为失败/取消；后续可继续加强日志与更严格的幂等/重复回调审计。

### P1：前端支付体验补齐
1. **已完成：展示二维码/跳转支付 + 状态轮询**：`createPaymentOrder` 返回 `pay_url/qr_url` 后，前端弹窗展示二维码/支付链接，并轮询订单状态（pending → paid/failed/expired），支付成功后刷新“我的订单”。
2. **已完成：支付结果页**：新增 `/payment/result` 页面，展示订单状态与到账提示；后端 return 回跳会携带 `order/status` 并跳转到该页面。
3. **已完成：支付方式约束（微信最低 ¥100）**：前端提示 + 按钮禁用已实现，后端创建订单也会对微信（Stripe）最低金额进行兜底校验。

### P2：配置与部署完善
1. **配置结构体补齐**：把实际需要的字段写入 `payment.zpay.*` / `payment.stripe.*`（如 submit/query/payment_methods/api_version/currency 等），并在 `.env.example` / `config.example.yaml` 里给出一致示例。
2. **已完成：环境变量兼容**：已兼容 `ZPAY_*`/`STRIPE_*` 前缀映射到 `payment.*`（除默认的 `PAYMENT_ZPAY_*`/`PAYMENT_STRIPE_*` 之外）。
3. **安全与排障**：补文档说明（回调域名必须 https、验签失败排查、IP 白名单等）。

## 1. 💰 支付系统
### 已落地
- `backend/internal/service/payment_service.go` 定义了订单创建、订单状态管理、额度发放以及与 `BalanceService`、`PromotionService`、`ReferralService` 的协作逻辑。
- `backend/internal/repository/payment_order_repo.go`、`backend/internal/repository/payment_cache.go` 以及 `backend/migrations/008_payment_order.sql` 已建立订单持久层与 Redis 速率限制/二维码缓存模型。
- `backend/internal/config/config.go` 已加入 `payment.*`（渠道开关、汇率、套餐、MaxOrders 等）。
- 已补齐用户端支付基础 API（用于让前端先跑通页面与数据展示）：
  - `GET /api/v1/payment/plans`（读取 `payment.packages` 输出套餐列表）
  - `POST /api/v1/payment/orders`（创建订单记录；需 `payment.enabled=true` 才允许创建）
  - `GET /api/v1/payment/orders`（查询我的订单列表，分页）
  - `GET /api/v1/payment/orders/:orderNo`（查询单个订单状态，用于前端支付弹窗轮询）
- 已补齐管理员侧充值记录（支付订单）查询/导出：
  - `GET /api/v1/admin/payment/orders`（分页 + 筛选：`method`/`user`）
  - `GET /api/v1/admin/payment/orders/export`（导出筛选后的记录，CSV）
  - 前端新增 `/admin/payment-orders` 页面，支持筛选、分页与“导出记录”。
- 已开始接入真实支付渠道（后端可返回 `pay_url/qr_url`）：
  - ZPay：`backend/internal/service/zpay_service.go` 生成收银台链接（`submit.php`）并支持回调签名校验。
  - Stripe：`backend/internal/service/stripe_service.go` 创建 `PaymentIntent`（微信支付），可返回 `HostedInstructionsURL` 与 QR 图（`image_url_png`）。
  - 回调路由已打通（无需认证）：`backend/internal/server/routes/payment.go`（ZPay notify / Stripe webhook）。
  - `POST /api/v1/payment/orders` 会根据 `provider` 返回 `pay_url/qr_url` 并写入订单 `payment_url` 字段。
- 前端已补齐 `/payment` 充值页（快速可见效果）：
  - 充值页样式/布局对齐 `docs/migrate-crs/zhifu.png`：固定 4 个套餐 + “其他（自定义金额）”，卡片展示 `$` 与“实付约 ¥xxx.xx”。
  - 选择卡片后高亮 + 右上角 ✅；点击后弹窗对齐 `docs/migrate-crs/tanchuang.png`，展示金额计算详情。
  - 实时计算公式：`CNY = USD × exchange_rate × discount_rate`（`discount_rate=0.15` 表示“支付 15%”）。
  - 支付方式在卡片下方展示为“支付宝 / 微信”（带图标），并在右侧提供“立即充值”按钮创建订单。
  - 兑换码相关入口在当前部署前端已注释（不作为充值路径）。
- 配置层已支持 `discount_rate`（支付倍率）与基于 `amount_usd` 的套餐配置；示例见 `deploy/config.example.yaml` 与 `deploy/.env.example`。
- 支付渠道字段已统一为 “支付宝/微信 → zpay/stripe” 映射：前端下单传 `channel=alipay|wechat`，后端内部归一化为 `provider=zpay|stripe`。
- 订单过期清理已接入后台任务：当 `payment.enabled=true` 时启动 `PaymentMaintenanceService` 定期调用 `PaymentService.CancelExpiredOrders`（见 `backend/internal/service/payment_maintenance_service.go`）。
- **已修复：`payment_orders.discount_rate` 非空约束导致下单失败**：订单创建时写入 `discount_rate`（支付倍率），并在 GORM 模型补齐 `DiscountRate` 字段，避免数据库已存在该列且无默认值时插入报错（`SQLSTATE 23502`）。
- **已补齐：旧环境变量兼容**：除 `PAYMENT_ZPAY_* / PAYMENT_STRIPE_*` 外，额外支持 `ZPAY_* / STRIPE_*` 变量名映射到 `payment.*` 配置（避免因环境变量命名不一致导致下单时报 “zpay/stripe is disabled”）。

### 待迁移
- ZPay 回调依赖公网可访问地址：若 `payment.zpay.notify_url/return_url` 为相对路径，需要配置 `payment.base_url`（否则后端无法拼接出完整回调 URL）。
- Stripe Webhook 目前已覆盖 `payment_intent.succeeded/payment_failed/canceled`，但仍可继续补齐更多事件（如退款）与更严格的对账校验/审计日志。

## 2. 🎁 活动优惠系统
### 已落地
- `backend/internal/service/promotion_service.go`、`internal/repository/promotion_repo.go`、`internal/repository/promotion_cache.go` 以及 `backend/migrations/006_promotion.sql` 构建了活动资格/记录存储。
- `AuthService.RegisterWithVerification` 会调用 `promotionService.InitUserPromotion`，`PaymentService.MarkOrderPaid` 会按档位计算赠送金额。
- `internal/config/config.go` 已暴露 `promotion.enabled/duration_hours/tiers` 配置。

### 待迁移
- 没有对外 API：路由/handler 中未出现 `promotion` 相关入口，用户/管理员无法查询活动状态或统计。
- 前端缺少 `PromotionBanner`、倒计时等组件，`frontend/src` 未存在任何 `promotion` 命名的文件，无法展示活动信息。

## 3. 👥 邀请返利系统
### 已落地
- `backend/internal/service/referral_service.go`、`internal/repository/referral_repo.go`、`internal/repository/referral_cache.go` 以及 `backend/migrations/007_referral.sql` 已实现邀请码/邀请关系/返利发放模型。
- `PaymentService.MarkOrderPaid` 会在充值成功时调用 `referralService.ProcessInviteeRecharge` 并在满足条件时向邀请人发放 `RechargeTypeReferral` 余额。
- `internal/config/config.go` 已提供 `referral.enabled/reward_usd/qualified_recharge_*` 配置。

### 待迁移
- 邀请关系无法写入：`RecordInvitation` 在任何 handler/服务中都未被调用，注册流程 (`internal/handler/auth_handler.go`) 也没有接收邀请码字段，导致 `referral_invites` 永远为空。
- 用户与管理员都没有访问接口或页面（Router 中没有 `/users/referral`，前端也没有邀请视图/组件）。
- 返利统计/邀请链接的生成逻辑缺失，`referralService.GetOrCreateUserCode` 未被任何入口使用。

## 4. 💵 用户余额系统
### 已落地
- `backend/internal/service/balance_service.go` + `internal/repository/recharge_record_repo.go` + `backend/migrations/005_recharge_record.sql` 已实现充值流水及扣减 API。
- 支付/返利路径会调用 `BalanceService.ApplyChange` 记账（见 `PaymentService` 中对 `RechargeTypePayment` 与 `RechargeTypeReferral` 的调用）。
- 兑换码充值已写入流水：`RedeemService` 的余额类兑换改为走 `BalanceService.ApplyChange`（用于数据一致性，前端当前不单独展示流水页）。
- **已补齐：后台充值写入账本**：管理员后台给用户加/减/设置余额会优先走 `BalanceService.ApplyChange` 写入 `recharge_records`（类型 `admin`），并在“加余额”场景额外创建一条 `payment_orders(provider=admin)` 以便用户侧“我的订单”可见“后台充值”。

### 待迁移
- 余额流水（`recharge_records`）仍缺少可用的管理员查询/导出 API；当前后台的“充值记录”页面使用的是 `payment_orders`（在线充值 + 后台充值），不等同于完整账本流水。
- 当前前端不单独提供“充值记录”页面：以“充值 → 我的订单”作为充值记录展示；如需展示完整余额流水，再补 `/user/recharge-records` 与对应页面。
- 若未来需要 `recharge_records` 层面的筛选/导出（按类型/日期/来源），需补相应 API 与后台页面/导出按钮。

## 5. 📧 邮件服务
### 已落地
- `backend/internal/service/email_service.go` 提供 SMTP 发送、HTML 模板、验证码生成、1 分钟冷却及 15 分钟 TTL；`internal/repository/email_cache.go` 用 Redis 存储验证码。
- `AuthService.SendVerifyCode(Async)`、`AuthHandler.SendVerifyCode`、`frontend/src/views/auth/EmailVerifyView.vue` 完成注册验证码流程。
- 管理端 `registerSettingsRoutes` 暴露 `/admin/settings/test-smtp`、`/send-test-email`，并通过 `SettingService` 读取 SMTP 配置。

### 待迁移
- CRS 中的密码重置邮件/邮箱验证 Token 未实现：仓库内没有 `password_reset`、reset token、`forgot password` 的服务或 API（`rg -ni "forgot"` 只在 README 中出现）。
- `docs/email-verification-password-reset.md` 等说明文档未迁移至 `docs/`。
- 前端缺少 `ForgotPasswordView.vue`、`ResetPasswordView.vue` 等页面。

## 6. 🔐 用户认证增强
### 已落地
- 本地注册/登录由 `AuthHandler.Register/Login` + JWT (`AuthService.GenerateToken`) + Turnstile 验证完成，`frontend/src/views/auth/RegisterView.vue`、`LoginView.vue` 已对应。
- 用户可通过 `PUT /api/v1/user/password` (`UserHandler.ChangePassword`) 修改密码，`internal/service/user.go`/`auth_service.go` 全面使用 bcrypt。
- 注册可配置默认余额/并发（`SettingService`）。

### 待迁移
- 缺失完整的忘记密码/重置口令流程（无 `POST /auth/forgot-password`、reset token、邮件模板及前端页面）。
- 旧 AES 密码自动迁移逻辑在当前实现中缺位，`AuthService` 只支持 bcrypt 哈希。
- 邮箱验证仅通过验证码而非带签名链接，无法覆盖 “邮箱验证页面” 和 “重置密码页面” 的需求。

## 7. 🔑 API Key 管理增强
### 已落地
- 现有 `ApiKeyService` 支持自定义 key、分组绑定、创建速率限制（`internal/repository/api_key_cache.go`）以及基础 CRUD。

### 待迁移
- 明文查看功能缺失：没有 `POST /admin/api-keys/reveal`，`internal/handler/admin` 内亦无相关 handler。
- 密钥仍以明文写入数据库（`internal/repository/api_key_repo.go` 的 `Key` 字段直接保存字符串），未见 AES-256-GCM 加密/解密流程。
- 未实现管理员口令验证、查看原因记录、`admin:reveal:audit` 审计日志或 Redis 速率限制。
- 统计接口缺失：`routes/admin.go` 中没有 `api-key-calls-metrics`，Redis 里也没有分钟级计数键。

## 8. 📊 用户管理优化
### 已落地
- `internal/repository/user_repo.go` 的 `ListWithFilters` 已包含状态/角色/关键字模糊检索，`frontend/src/views/admin/UsersView.vue` 支持筛选与余额调整。
- 已补齐筛选导出：`GET /api/v1/admin/users/export` 导出当前筛选条件下的全部用户记录（前端按钮名为“导出记录”）。

### 待迁移
- 用户列表/统计未做缓存：仓库中没有 `user:list`、`user:stats` 类 Redis 缓存，也没有 TTL 设置。
- 并发控制/分页优化缺失，代码中未使用 changelog 中提及的 `p-limit` 或类似并发限制库，仍是直接 DB 查询。
- `UserManagementView`、`UserDashboardView` 等页面没有体现 changelog 中的 UI/UX 调整（禁用分页、增强图表等）。

## 9. 🛡️ 鉴权检测安全修复
### 状态
- CRS 修复针对 Redis session 伪造（参考 `lei-claude-relay-service` commit `0eef7dcd` 中新增的 `cleanupInvalidSessions` 和 session 字段校验）。
- Sub2API 使用 JWT + Admin API Key (`backend/internal/server/middleware/admin_auth.go`)，不会在 Redis 存储可伪造的 session，因而该漏洞路径不存在，暂无需迁移。
- 建议在文档中记录此差异，若未来引入服务端 session，再回溯该修复。

## 10. 🖥️ 前端界面改造
### 已落地（本仓库新增）
- 新增用户端路由与页面：
  - `/payment`：`frontend/src/views/user/PaymentView.vue` + `frontend/src/router/index.ts`
- 已将入口补到侧边栏导航（用户与管理员“我的账户”区）：`frontend/src/components/layout/AppSidebar.vue`（新增 `nav.payment`）
- 用户仪表盘 “快捷操作” 增加了前往充值入口：`frontend/src/views/user/DashboardView.vue`
 - 预置 API 模块（用于后续对接后端；当后端 404 时前端给出“接口未启用”提示，不会白屏）：`frontend/src/api/payment.ts`
- i18n 文案补齐：`frontend/src/i18n/locales/en.ts`、`frontend/src/i18n/locales/zh.ts`
- 兑换码相关入口在当前部署前端已注释（路由/侧边栏/快捷入口均隐藏）。

### 现状
- `frontend/src/views` 主要包含 Dashboard/Keys/Usage/Profile/Subscriptions 及 admin 视图；当前已新增 `/payment`（充值套餐+订单列表），Redeem 相关页面虽存在但在此部署默认不作为入口暴露。
- `frontend/src/components`、`src/views` 下不存在 `PromotionBanner.vue`、`UserManualView.vue`、`UserRechargeRecords.vue`、`ContactUsModal.vue`、`ConfigurationGuideModal.vue`、`EnvironmentSetupGuide.vue`、`PlatformCodeSnippet.vue`、`PasswordStrengthMeter.vue`、`RevealApiKeyModal.vue` 等文件。
- `/payment` 的核心 UI/UX（选择套餐、弹窗计算、支付方式选择、立即充值）已落地；仍待对接真实支付渠道后补齐支付链接/二维码展示与支付完成体验。

### 待迁移
- 根据 changelog 新增的 SPA 组件/页面需要在 Vue 层逐一补齐，并与未来 `/payment` API 打通。

## 11. 🤖 钉钉机器人集成
### 状态
- 仓库搜索 `rg -ni "dingtalk" -n` 无任何结果，`src/routes/dingtalkBot.js` 对应的后端路由尚未迁移。
- 若仍需在 Sub2API 中隐藏充值操作员，需要新增 Go handler、配置项及可能的 webhook。

## 12. 🏗️ 架构调整
### 现状
- Sub2API 已是前后端分离的 Go + Vue 仓库，但 changelog 中的支撑文件未出现：
  - 仅存在 `deploy/docker-compose.yml`，没有 `docker-compose-dev.yml`、`docker-compose.repo.yml`。
  - 根目录没有 `crs-compose.sh`、`setup-docker-compose.sh`。
  - `docs/` 下也没有 `nginx.example.conf` 与 `ROUTING.md`。

### 待迁移
- 根据需要补充开发/仓库版 Compose、脚本与反向代理示例，或在文档中注明 Sub2API 的等效方案。

## 13. 📝 其他改进
### 缺失项
- 无 `scripts/` 目录，更没有 `generate-test-usage-data.js`、`migrate-add-user-apikey-index.js`、`migrate-user-authtype.js` 等工具。
- `frontend/src/composables` 中没有 `useEnvironmentConfig.js`，也未发现 clipboard 工具或输入校验增强。
- 文档 `USER_API_REFERENCE.md`、`docs/stripe-payment-analysis.md`、`docs/user-balance-payment.md` 未迁移到 `docs/`。

## Redis Key 规划（Sub2API 适配）

### 设计原则
- **优先 PostgreSQL**：订单、邀请、流水等持久化数据仅保留在数据库中，Redis 只承担速率限制、缓存或短期令牌，避免状态漂移。
- **统一命名空间**：采用 `模块:用途:{标识}` 格式，与现有 `verify_code:*`、`payment:*`、`billing:*` 等前缀风格一致。
- **明确 TTL**：所有易失键均指定 TTL，非易失键仅在确有需要时常驻，并配套后台巡检。

### 变更对照（基于 CRS 列表）
| 功能 | 原 CRS Key | Sub2API 方案 | 状态 | 说明 |
|------|------------|--------------|------|------|
| 支付订单缓存 | `payment_order:{orderId}` / `payment_orders_user:{userId}` / `payment_orders_all` | 直接使用 PostgreSQL `payment_orders` 表（`backend/migrations/008_payment_order.sql`）+ GORM 仓储；Redis 仅保留 `payment:url:{orderNo}`（现有，TTL=订单到期时间，用于支付链接）和 `payment:counter:{userId}`（现有，TTL=1 分钟，用于速率限制）。 | ✅ 已替代 | 避免订单双写，沿用现有实现。 |
| 活动资格 | `user_promotion:{userId}` | 复用 `promotion:user:{userId}`（已在 `backend/internal/repository/promotion_cache.go` 实现，TTL=资格剩余时间）。 | ✅ 已存在 | --- |
| 活动统计 | `promotion_stats:*` | 计划新增 `promotion:stats:{yyyymmdd}`（Hash，记录 `created`/`used` 等聚合值，TTL=3 天），由后台定时器写入，管理端直接读取。 | ⏳ 待新增 | 仅影响活动报表，无需阻塞主流程。 |
| 邀请码映射 | `referral:code:{code}` | 沿用 `referral:code:{code}`（现有，TTL=24 小时，用于 code → userID）。 | ✅ 已存在 | --- |
| 用户→邀请码 | `referral:user:{userId}:code` | 新增 `referral:user_code:{userId}`（String，TTL=24 小时），`ReferralService.GetOrCreateUserCode` 保存/回源，减少频繁 SQL。 | ⏳ 待新增 | 在上线邀请入口前实施。 |
| 邀请缓存 | `referral:invite:{inviteeId}` | 新增 `referral:invite_cache:{inviteeId}`（JSON，TTL=24 小时）缓存 `referral_invites` 行数据，供 invitee 查询状态；写路径仍以数据库为准。 | ⏳ 待新增 | 用于高频读取。 |
| 邀请统计 | `referral:stats:{userId}` | 新增 `referral:stats:{userId}`（Hash，字段 `total`/`qualified`/`rewarded`，TTL=10 分钟），后台在写操作后失效或异步刷新。 | ⏳ 待新增 | 保障排行榜/仪表盘性能。 |
| API Key 缓存 | `user_apikeys:{userId}` | API Key 元数据存储在 `api_keys` 表；Redis 仅用于：①既有 `apikey:ratelimit:{userId}`（限制自定义 Key 错误次数，TTL=24 小时）；②新增 `apikey:calls:{apiKey}:{yyyyMMddHHmm}`（Hash 或 String，TTL=2 小时）记录分钟级调用量以支撑 `GET /admin/api-key-calls-metrics`。 | ⏳ 部分新增 | 不复制整表，仅缓存统计。 |
| 密码重置 | `password_reset_token:{hash}` | 新增 `auth:password_reset:{token}`（String→userID，TTL=30 分钟），配合即将实现的忘记密码/重置接口，在成功重置后立即删除。 | ⏳ 待新增 | 依赖邮件服务。 |
| 邮箱验证 | `email_verification_token:{hash}` | 现有 `verify_code:{email}`（见 `backend/internal/repository/email_cache.go`，TTL=15 分钟）满足验证码式验证；如需链接式验证，再补充 `auth:email_verify:{token}`（String→email，TTL=30 分钟）。 | ✅/⏳ | 视 UI 方案决定。 |
| 充值流水缓存 | `recharge_record:{recordId}` | 维持 PostgreSQL `recharge_records` 表为唯一来源；若需优化列表，可新增 `recharge:list_cache:{userId}`（JSON，TTL=5 分钟）做分页缓存，但不做记录级镜像。 | ✅ 已替代 | 与记账一致。 |
| 管理员查看审计 | `admin:reveal:audit` | 新增 `audit:api_key_reveal`（List，元素为 JSON：`adminId`/`apiKeyId`/`reason`/`ts`，不设 TTL 仅保留最近 N 条）及 `audit:api_key_reveal_rate:{adminId}`（String，TTL=1 分钟）做速率保护。 | ⏳ 待新增 | 配合“API Key 明文查看”上线。 |

### 现有 Redis 前缀参考
- `verify_code:{email}`：邮件验证码（TTL=15 分钟）。
- `payment:counter:{userId}`：支付下单速率限制（TTL=1 分钟，见 `payment_cache.go`）。
- `payment:url:{orderNo}`：支付链接缓存（TTL=订单到期）。
- `billing:balance:{userId}` / `billing:sub:{userId}:{groupId}`：余额、订阅缓存（TTL=5 分钟）。
- `promotion:user:{userId}`、`referral:code:{code}`、`redeem:ratelimit:{userId}`、`apikey:ratelimit:{userId}` 等：均已在 `backend/internal/repository` 下实现，可作为新 key 命名与 TTL 的参考。

---

如需推进迁移，可按上述待办逐条建 issue 或纳入 sprint；完成后同步更新本文件。
