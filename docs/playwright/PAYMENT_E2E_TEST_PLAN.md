# 支付系统 E2E 测试计划 (Playwright MCP)

基于 Playwright MCP 的端到端测试流程，覆盖支付、活动优惠、邀请返利等核心功能。

---

## 一、测试环境准备

### 1.1 前置条件

```yaml
环境要求:
  - 后端服务已启动 (localhost:8080)
  - 前端服务已启动 (localhost:5173)
  - 数据库已初始化
  - Redis 已启动
  - Zpay 测试环境配置完成

测试账号:
  - 管理员: admin@test.com / admin123
  - 新用户: 测试时动态注册
  - 邀请人: inviter@test.com / test123
```

### 1.2 测试数据重置

```
每个测试套件开始前:
1. 清理测试用户数据
2. 重置活动配置为默认值
3. 清理测试订单数据
```

---

## 二、测试用例

### 2.1 用户注册与活动初始化

#### TC-001: 新用户注册自动初始化活动

```
测试目标: 验证新用户注册后自动获得 72 小时活动资格

步骤:
1. 打开注册页面 /register
2. 填写邮箱: test_{{timestamp}}@test.com
3. 填写密码: Test123456
4. 点击注册按钮
5. 等待跳转到用户首页

验证:
- [x] 注册成功，跳转到首页
- [x] 页面显示活动横幅 "新人限时优惠"
- [x] 活动横幅显示剩余时间 (约 72 小时)
- [x] 活动横幅显示当前可获得 30% 返赠
```

#### TC-002: 使用邀请码注册

```
测试目标: 验证使用邀请码注册建立邀请关系

前置:
- 邀请人 inviter@test.com 已登录并获取邀请码

步骤:
1. 邀请人登录，进入邀请页面 /user/referral
2. 复制邀请码 (如: ABC12345)
3. 新窗口打开注册页面 /register?code=ABC12345
4. 填写邮箱: invitee_{{timestamp}}@test.com
5. 填写密码: Test123456
6. 点击注册

验证:
- [x] 注册成功
- [x] 邀请人的邀请页面显示新增被邀请人
- [x] 被邀请人状态显示 "待充值"
```

---

### 2.2 充值流程测试

#### TC-003: 创建充值订单

```
测试目标: 验证创建充值订单流程

前置:
- 用户已登录

步骤:
1. 进入充值页面 /user/payment
2. 选择充值金额 ¥100
3. 选择支付方式: 支付宝
4. 点击 "立即充值"
5. 等待生成订单

验证:
- [x] 显示订单信息弹窗
- [x] 订单金额显示 ¥100
- [x] 显示预计到账金额 (含活动返赠)
- [x] 显示支付二维码或跳转链接
- [x] 显示订单倒计时 (30 分钟)
```

#### TC-004: 支付成功回调处理

```
测试目标: 验证支付成功后余额更新

前置:
- 已创建待支付订单
- Zpay 测试环境配置模拟回调

步骤:
1. 触发 Zpay 模拟支付成功回调
2. 刷新用户页面
3. 检查余额变化

验证:
- [x] 余额增加正确金额 (本金 + 活动返赠)
- [x] 订单状态变为 "已支付"
- [x] 充值记录页面显示新记录
- [x] 活动横幅消失或显示 "已使用"
```

#### TC-005: 订单超时自动取消

```
测试目标: 验证超时订单自动取消

步骤:
1. 创建充值订单
2. 等待 31 分钟 (或手动触发过期任务)
3. 刷新订单列表

验证:
- [x] 订单状态变为 "已过期"
- [x] 用户余额未变化
- [x] 活动资格未消耗
```

---

### 2.3 活动优惠测试

#### TC-006: 24小时内充值享 30% 返赠

```
测试目标: 验证新用户 24 小时内充值获得 30% 返赠

前置:
- 新注册用户 (注册时间 < 24 小时)

步骤:
1. 登录新用户账号
2. 进入充值页面
3. 充值 ¥100 (约 $13.89)
4. 完成支付

验证:
- [x] 到账金额 = $13.89 + $4.17 (30%) = $18.06
- [x] 充值记录显示 "活动返赠: $4.17"
- [x] 活动状态变为 "已使用"
```

#### TC-007: 活动层级递减验证

```
测试目标: 验证不同时间段活动返赠比例

测试矩阵:
| 注册后时间 | 预期返赠 |
|-----------|---------|
| 12 小时   | 30%     |
| 30 小时   | 20%     |
| 40 小时   | 10%     |
| 60 小时   | 5%      |
| 80 小时   | 0% (已过期) |

步骤:
1. 修改用户注册时间 (数据库)
2. 刷新充值页面
3. 检查显示的返赠比例

验证:
- [x] 各时间段返赠比例正确
- [x] 过期后不显示活动横幅
```

#### TC-008: 活动只能使用一次

```
测试目标: 验证活动优惠只能使用一次

前置:
- 用户已使用过活动优惠

步骤:
1. 登录已使用活动的用户
2. 进入充值页面
3. 再次充值

验证:
- [x] 不显示活动横幅
- [x] 充值预览不包含返赠
- [x] 实际到账金额 = 充值金额 / 汇率 (无返赠)
```

---

### 2.4 邀请返利测试

#### TC-009: 生成邀请码

```
测试目标: 验证用户可以生成唯一邀请码

步骤:
1. 登录用户账号
2. 进入邀请页面 /user/referral
3. 查看邀请码

验证:
- [x] 显示 8 位邀请码
- [x] 显示邀请链接
- [x] 复制按钮可用
- [x] 多次刷新邀请码不变
```

#### TC-010: 被邀请人充值达标触发返利

```
测试目标: 验证被邀请人充值达标后邀请人获得返利

前置:
- 邀请人 A 已邀请用户 B
- 用户 B 累计充值 < ¥20

步骤:
1. 用户 B 登录
2. 充值 ¥20
3. 完成支付
4. 邀请人 A 刷新页面

验证:
- [x] 用户 B 充值成功
- [x] 邀请人 A 余额增加 $10
- [x] 邀请人 A 充值记录显示 "邀请返利"
- [x] 邀请列表中用户 B 状态变为 "已达标"
```

#### TC-011: 返利只发放一次

```
测试目标: 验证同一被邀请人只触发一次返利

前置:
- 用户 B 已达标，邀请人 A 已获得返利

步骤:
1. 用户 B 再次充值 ¥50
2. 检查邀请人 A 余额

验证:
- [x] 邀请人 A 余额不变
- [x] 不产生新的返利记录
```

---

### 2.5 余额与记录测试

#### TC-012: 充值记录列表

```
测试目标: 验证充值记录正确显示

步骤:
1. 进入充值页面 /payment（“我的订单”即充值记录）
2. 查看记录列表

验证:
- [x] 显示所有充值/扣款记录
- [x] 记录包含: 时间、类型、金额、余额变化
- [x] 支持分页
- [x] 按时间倒序排列
```

#### TC-013: 余额记录类型区分

```
测试目标: 验证不同类型记录正确标识

验证记录类型:
| 类型 | 来源 | 金额方向 |
|------|------|---------|
| payment | 在线充值 | + |
| admin | 管理员调整 | +/- |
| redeem | 兑换码 | + |
| referral | 邀请返利 | + |
| deduct | API 扣费 | - |

验证:
- [x] 各类型记录图标/颜色区分
- [x] 金额正负显示正确
```

---

### 2.6 管理后台测试

#### TC-014: 管理员查看订单列表

```
测试目标: 验证管理员可以查看所有订单

步骤:
1. 管理员登录
2. 进入订单管理 /admin/payment/orders
3. 查看订单列表

验证:
- [x] 显示所有用户订单
- [x] 支持按状态筛选
- [x] 支持按时间范围筛选
- [x] 显示订单详情
```

#### TC-015: 管理员手动调整余额

```
测试目标: 验证管理员可以手动调整用户余额

步骤:
1. 管理员进入用户管理
2. 选择目标用户
3. 点击 "调整余额"
4. 输入金额和备注
5. 确认操作

验证:
- [x] 用户余额更新正确
- [x] 产生 admin 类型充值记录
- [x] 记录包含操作者和备注
```

---

## 三、Playwright MCP 测试脚本

### 3.1 测试目录结构

```
tests/
├── e2e/
│   ├── payment/
│   │   ├── create-order.spec.ts
│   │   ├── payment-callback.spec.ts
│   │   └── order-timeout.spec.ts
│   ├── promotion/
│   │   ├── new-user-bonus.spec.ts
│   │   └── tier-decay.spec.ts
│   ├── referral/
│   │   ├── generate-code.spec.ts
│   │   ├── invite-reward.spec.ts
│   │   └── reward-once.spec.ts
│   ├── balance/
│   │   ├── (removed) recharge-records.spec.ts
│   │   └── record-types.spec.ts
│   └── admin/
│       ├── order-management.spec.ts
│       └── balance-adjustment.spec.ts
├── fixtures/
│   ├── auth.ts
│   └── test-data.ts
└── playwright.config.ts
```

### 3.2 示例测试脚本

#### 新用户注册活动测试

```typescript
// tests/e2e/promotion/new-user-bonus.spec.ts

import { test, expect } from '@playwright/test';

test.describe('新用户活动优惠', () => {

  test('TC-001: 新用户注册自动初始化活动', async ({ page }) => {
    const timestamp = Date.now();
    const email = `test_${timestamp}@test.com`;

    // 1. 打开注册页面
    await page.goto('/register');

    // 2. 填写注册表单
    await page.fill('[data-testid="email-input"]', email);
    await page.fill('[data-testid="password-input"]', 'Test123456');
    await page.fill('[data-testid="confirm-password-input"]', 'Test123456');

    // 3. 点击注册
    await page.click('[data-testid="register-button"]');

    // 4. 等待跳转
    await page.waitForURL('/dashboard');

    // 5. 验证活动横幅
    const banner = page.locator('[data-testid="promotion-banner"]');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText('新人限时优惠');
    await expect(banner).toContainText('30%');
  });

  test('TC-006: 24小时内充值享30%返赠', async ({ page }) => {
    // 前置: 登录新注册用户
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'newuser@test.com');
    await page.fill('[data-testid="password-input"]', 'Test123456');
    await page.click('[data-testid="login-button"]');
    await page.waitForURL('/dashboard');

    // 1. 进入充值页面
    await page.goto('/user/payment');

    // 2. 选择充值金额
    await page.click('[data-testid="amount-100"]');

    // 3. 验证预览信息
    const preview = page.locator('[data-testid="payment-preview"]');
    await expect(preview).toContainText('¥100');
    await expect(preview).toContainText('活动返赠');
    await expect(preview).toContainText('30%');

    // 4. 验证预计到账金额
    // ¥100 / 7.2 = $13.89, 返赠 30% = $4.17, 总计 $18.06
    await expect(preview).toContainText('$18.06');
  });
});
```

#### 邀请返利测试

```typescript
// tests/e2e/referral/invite-reward.spec.ts

import { test, expect } from '@playwright/test';

test.describe('邀请返利', () => {

  test('TC-010: 被邀请人充值达标触发返利', async ({ browser }) => {
    // 创建两个浏览器上下文 (邀请人 + 被邀请人)
    const inviterContext = await browser.newContext();
    const inviteePage = await inviterContext.newPage();

    const inviteeContext = await browser.newContext();
    const inviterPage = await inviteeContext.newPage();

    // === 邀请人操作 ===
    // 1. 邀请人登录
    await inviterPage.goto('/login');
    await inviterPage.fill('[data-testid="email-input"]', 'inviter@test.com');
    await inviterPage.fill('[data-testid="password-input"]', 'test123');
    await inviterPage.click('[data-testid="login-button"]');

    // 2. 获取初始余额
    await inviterPage.goto('/user/referral');
    const initialBalance = await inviterPage.locator('[data-testid="user-balance"]').textContent();

    // 3. 获取邀请码
    const inviteCode = await inviterPage.locator('[data-testid="invite-code"]').textContent();

    // === 被邀请人操作 ===
    // 4. 被邀请人使用邀请码注册
    await inviteePage.goto(`/register?code=${inviteCode}`);
    await inviteePage.fill('[data-testid="email-input"]', `invitee_${Date.now()}@test.com`);
    await inviteePage.fill('[data-testid="password-input"]', 'Test123456');
    await inviteePage.click('[data-testid="register-button"]');
    await inviteePage.waitForURL('/dashboard');

    // 5. 被邀请人充值 ¥20 (模拟支付成功)
    await inviteePage.goto('/user/payment');
    await inviteePage.fill('[data-testid="custom-amount"]', '20');
    await inviteePage.click('[data-testid="pay-button"]');

    // 模拟支付回调 (通过 API 或测试后门)
    await inviteePage.evaluate(async () => {
      await fetch('/api/test/simulate-payment-callback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ orderNo: 'latest' })
      });
    });

    // === 验证邀请人获得返利 ===
    // 6. 邀请人刷新页面
    await inviterPage.reload();

    // 7. 验证余额增加 $10
    const newBalance = await inviterPage.locator('[data-testid="user-balance"]').textContent();
    const balanceDiff = parseFloat(newBalance!) - parseFloat(initialBalance!);
    expect(balanceDiff).toBeCloseTo(10, 1);

    // 8. 验证邀请列表状态
    const inviteeStatus = inviterPage.locator('[data-testid="invitee-status"]').first();
    await expect(inviteeStatus).toContainText('已达标');

    // 清理
    await inviterContext.close();
    await inviteeContext.close();
  });
});
```

#### 充值订单测试

```typescript
// tests/e2e/payment/create-order.spec.ts

import { test, expect } from '@playwright/test';

test.describe('充值订单', () => {

  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login');
    await page.fill('[data-testid="email-input"]', 'testuser@test.com');
    await page.fill('[data-testid="password-input"]', 'test123');
    await page.click('[data-testid="login-button"]');
    await page.waitForURL('/dashboard');
  });

  test('TC-003: 创建充值订单', async ({ page }) => {
    // 1. 进入充值页面
    await page.goto('/user/payment');

    // 2. 选择充值金额
    await page.click('[data-testid="amount-100"]');

    // 3. 选择支付方式
    await page.click('[data-testid="payment-method-alipay"]');

    // 4. 点击充值
    await page.click('[data-testid="create-order-button"]');

    // 5. 等待订单弹窗
    const orderModal = page.locator('[data-testid="order-modal"]');
    await expect(orderModal).toBeVisible();

    // 6. 验证订单信息
    await expect(orderModal).toContainText('¥100');
    await expect(orderModal).toContainText('支付宝');

    // 7. 验证二维码或支付链接存在
    const qrCode = orderModal.locator('[data-testid="payment-qrcode"]');
    const payLink = orderModal.locator('[data-testid="payment-link"]');
    const hasPayment = await qrCode.isVisible() || await payLink.isVisible();
    expect(hasPayment).toBeTruthy();

    // 8. 验证倒计时
    const countdown = orderModal.locator('[data-testid="order-countdown"]');
    await expect(countdown).toBeVisible();
  });

  test('TC-005: 订单超时自动取消', async ({ page }) => {
    // 1. 创建订单
    await page.goto('/user/payment');
    await page.click('[data-testid="amount-10"]');
    await page.click('[data-testid="create-order-button"]');

    // 2. 获取订单号
    const orderNo = await page.locator('[data-testid="order-no"]').textContent();

    // 3. 手动触发过期任务 (测试环境 API)
    await page.evaluate(async () => {
      await fetch('/api/test/expire-orders', { method: 'POST' });
    });

    // 4. 查看订单状态
    await page.goto('/user/payment/orders');
    const orderRow = page.locator(`[data-testid="order-${orderNo}"]`);

    // 5. 验证状态为已过期
    await expect(orderRow.locator('[data-testid="order-status"]')).toContainText('已过期');
  });
});
```

### 3.3 Playwright 配置

```typescript
// playwright.config.ts

import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',

  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: [
    {
      command: 'cd backend && go run main.go',
      url: 'http://localhost:8080/health',
      reuseExistingServer: !process.env.CI,
    },
    {
      command: 'cd frontend && npm run dev',
      url: 'http://localhost:5173',
      reuseExistingServer: !process.env.CI,
    },
  ],
});
```

---

## 四、测试执行

### 4.1 执行命令

```bash
# 运行所有 E2E 测试
npx playwright test

# 运行特定测试文件
npx playwright test tests/e2e/payment/create-order.spec.ts

# 运行特定测试套件
npx playwright test --grep "新用户活动"

# 带 UI 运行 (调试模式)
npx playwright test --ui

# 生成测试报告
npx playwright show-report
```

### 4.2 MCP 执行流程

```
使用 Playwright MCP 执行测试:

1. 启动测试环境
   > mcp: playwright_navigate url="http://localhost:5173"

2. 执行注册流程
   > mcp: playwright_fill selector="[data-testid='email-input']" value="test@test.com"
   > mcp: playwright_fill selector="[data-testid='password-input']" value="Test123456"
   > mcp: playwright_click selector="[data-testid='register-button']"

3. 验证活动横幅
   > mcp: playwright_screenshot name="after-register"
   > mcp: playwright_evaluate script="document.querySelector('[data-testid=\"promotion-banner\"]').textContent"

4. 执行充值流程
   > mcp: playwright_navigate url="http://localhost:5173/user/payment"
   > mcp: playwright_click selector="[data-testid='amount-100']"
   > mcp: playwright_click selector="[data-testid='create-order-button']"
   > mcp: playwright_screenshot name="payment-order-created"
```

---

## 五、测试检查清单

### 5.1 冒烟测试 (每次部署)

| # | 测试项 | 优先级 |
|---|-------|-------|
| 1 | 用户注册 | P0 |
| 2 | 用户登录 | P0 |
| 3 | 创建充值订单 | P0 |
| 4 | 支付回调处理 | P0 |
| 5 | 余额更新 | P0 |

### 5.2 回归测试 (每周)

| # | 测试项 | 优先级 |
|---|-------|-------|
| 1 | 所有注册场景 | P1 |
| 2 | 活动各层级验证 | P1 |
| 3 | 邀请返利全流程 | P1 |
| 4 | 订单超时处理 | P1 |
| 5 | 管理后台功能 | P2 |

### 5.3 性能测试 (每月)

| # | 测试项 | 指标 |
|---|-------|------|
| 1 | 充值页面加载 | < 2s |
| 2 | 订单创建响应 | < 500ms |
| 3 | 支付回调处理 | < 1s |
| 4 | 并发订单创建 | 100 QPS |
