# 支付启用逻辑血缘备注

这份备注只用于识别当前代码里两套支付启用逻辑的来源，方便后续排查 `payment_enabled` 行为不一致的问题。

## 1. 旧版 config/env 支付体系

识别方式：

- Go 配置结构：`config.PaymentConfig`
- 配置键：`payment.enabled`
- 环境变量：`PAYMENT_ENABLED`
- 渠道配置：`payment.zpay.enabled`、`payment.zpay.payment_methods`、`payment.stripe.enabled`
- 典型判断：`cfg.Payment.Enabled` 或 `s.cfg.Enabled`

主要来源：

- 作者：`zhaoyapeng <zhaoyapeng@laiye.ai>`
- 首次提交：`284ff0a7d9415cd2c38b44c1b7b8e014b615d196`
- 提交信息：`支付系统第一版`
- 关键文件：
  - `backend/internal/config/config.go`
  - `backend/internal/service/payment_service.go`

后续重要改动：

- `a525189a5`，作者 `zhaoyapeng`：绑定 `PAYMENT_ENABLED`、`PAYMENT_ZPAY_ENABLED`、`PAYMENT_STRIPE_ENABLED` 等环境变量。
- `4cfe1559`，作者 `Claude Code`：增加用户侧 `alipay` / `wechat` 渠道开关识别，读取 `payment.zpay.payment_methods`。
- `82c10bd0`，作者 `Claude Code`：公开设置里的 `payment_enabled` 改为读取 `cfg.Payment.Enabled`，用于隐藏支付入口。

当前典型路径：

- `/api/v1/settings/public` 返回的 `payment_enabled` 来自 `cfg.Payment.Enabled`。
- 旧版 `/api/v1/payment/orders` 用户充值下单会检查 `s.cfg.Enabled`。
- 旧版用户侧支付宝 / 微信展示和下单还会检查 `payment.zpay.enabled` 与 `payment.zpay.payment_methods`。

## 2. 新版 DB settings/provider_instances 支付体系

识别方式：

- 服务：`PaymentConfigService`
- DB settings key：`payment_enabled`
- 常量：`service.SettingPaymentEnabled`
- Provider 表：`payment_provider_instances`
- 实例开关：`payment_provider_instances.enabled`
- 典型判断：`cfg.Enabled`，其中 `cfg` 来自 `PaymentConfigService.GetPaymentConfig`

主要来源：

- 作者：`erio <asakifeng@gmail.com>`
- 首次提交：`63d1860dc0063df3da020d0c638288af469ebb65`
- 提交信息：`feat(payment): add complete payment system with multi-provider support`
- 合入方式：PR `#1572`，来源分支 `touwaeriol/feat/payment-system-v2`
- Merge commit：`97f14b7a086bf75c72b3549e0d546907a720eb8e`
- 关键文件：
  - `backend/internal/service/payment_config_service.go`
  - `backend/internal/service/payment_order.go`
  - `backend/ent/schema/payment_provider_instance.go`

后续重要改动：

- `IanShaw027`：围绕支付回跳、微信 JSAPI、resume token、webhook 路由做过多次修复，例如 `b51bc7ee`、`7ef7fd19`、`1d8432b8`。
- `0c162a0d`，作者 `Claude Code`：恢复 `/api/v1/payment/checkout-info` 兼容接口。

当前典型路径：

- 新版 `CreateOrderV119` 会读取 `PaymentConfigService.GetPaymentConfig`。
- `PaymentConfig.Enabled` 来自 settings 表里的 `payment_enabled`。
- 实际支付方式是否可用还依赖 `payment_provider_instances.enabled`、`provider_key`、`supported_types`。

## 3. 容易混淆的点

当前代码里同时存在两个名字相近的启用开关：

- `config.PaymentConfig.Enabled`：旧版 config/env 体系，来自 `payment.enabled` / `PAYMENT_ENABLED`。
- `service.PaymentConfig.Enabled`：新版 DB settings 体系，来自 settings 表 `payment_enabled`。

因此看到 `payment_enabled` 时，需要先看它的来源：

- 如果来自 `/api/v1/settings/public` 或 `cfg.Payment.Enabled`，属于旧版 config/env 体系，源头是 `zhaoyapeng` 的第一版支付系统。
- 如果来自 `PaymentConfigService.GetPaymentConfig` 或 `SettingPaymentEnabled`，属于新版 DB-backed payment-system-v2，源头是 `erio` 的多 provider 支付系统。
