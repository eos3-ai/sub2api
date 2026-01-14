# Sub2API Playwright E2E（Demo / Prod）

本目录用于对已部署的 Sub2API 前端站点做端到端验证（Playwright）。

## 1) 配置（不进 Git）

1. 复制示例配置：
   - `cp e2e.config.example.json e2e.config.local.json`
2. 编辑 `e2e.config.local.json`，填写：
   - `baseURL`
   - 普通用户/管理员账号密码
   - 注册用的 `registration.password` 与 `registration.emailDomain`

说明：
- `e2e.config.local.json` 在仓库根 `.gitignore` 中已忽略，不会被提交。

## 2) 安装依赖

在 `tests/e2e/` 目录下执行：

- `npm install`
- 安装浏览器（首次需要）：`npx playwright install`

## 3) 运行命令

- Demo 环境：`npm run test:demo`
- Prod 环境：`npm run test:prod`

等价于：
- `cd tests/e2e && npm run test:demo`

## 4) 覆盖用例（对应 checklist）

- 注册账户 + 活动页展示：`tests/e2e/tests/01-register-and-promotion.spec.ts`
- 两种方式充值（支付宝/微信）：`tests/e2e/tests/02-recharge-two-methods.spec.ts`
- 邀请链接注册 + 发起充值（验证邀请关系建立）：`tests/e2e/tests/03-referral-invite.spec.ts`
- API 密钥创建删除：`tests/e2e/tests/04-api-keys-crud.spec.ts`
- 使用记录页可访问：`tests/e2e/tests/05-usage-records.spec.ts`

## 5) 自动化边界

当前脚本覆盖到：
- 创建订单成功（弹窗出现、订单号可读、二维码存在/或提示无二维码）
- 新人首充优惠模块展示（倒计时/文案）
- 邀请关系建立（邀请统计 +1）

但以下环节依赖真实支付成功回调（或专用测试回调接口），通常无法在纯前端 E2E 中稳定自动完成：
- 余额到账
- 活动赠送额度真正入账
- 邀请返利真正发放

如果 demo 环境具备“模拟支付成功”的能力（比如后台按钮或测试 API），我可以继续补齐“支付成功 → 余额变化 → 邀请返利发放”的自动化流程。

