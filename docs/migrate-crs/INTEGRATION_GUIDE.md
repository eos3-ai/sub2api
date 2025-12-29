# 功能模块整合到其他项目指南

本文档详细说明如何将各功能模块整合到其他项目中，包括模块依赖分析、文件清单、整合顺序和适配要点。

---

## 目录

1. [功能模块概览](#1-功能模块概览)
2. [模块依赖关系图](#2-模块依赖关系图)
3. [基础依赖层（必须先整合）](#3-基础依赖层必须先整合)
4. [功能模块详解](#4-功能模块详解)
5. [推荐整合顺序](#5-推荐整合顺序)
6. [整合方案选择](#6-整合方案选择)
7. [适配改造要点](#7-适配改造要点)
8. [接口对接说明](#8-接口对接说明)

---

## 1. 功能模块概览

| 模块 | 独立性 | 依赖复杂度 | 代码量 | 移植难度 |
|------|--------|------------|--------|----------|
| 📧 邮件服务 | ⭐⭐⭐⭐⭐ 完全独立 | 低 | ~330行 | 简单 |
| 🎁 活动优惠 | ⭐⭐⭐⭐ 基本独立 | 低 | ~580行 | 简单 |
| 👥 邀请返利 | ⭐⭐⭐⭐ 基本独立 | 低 | ~540行 | 简单 |
| 💵 用户余额 | ⭐⭐⭐ 需适配 | 中 | ~1200行 | 中等 |
| 💰 支付系统 | ⭐⭐ 依赖多 | 高 | ~1600行 | 复杂 |
| 🔐 本地认证 | ⭐⭐⭐ 需适配 | 中 | ~800行 | 中等 |

---

## 2. 模块依赖关系图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          基础依赖层 (必须)                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐       │
│  │  Redis  │  │  Logger │  │  Config │  │ crypto  │  │  bcrypt │       │
│  │  模型   │  │  工具   │  │  配置   │  │ (内置)  │  │  (npm)  │       │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘       │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│   📧 邮件服务  │          │  🎁 活动优惠  │          │  👥 邀请返利   │
│  EmailService │          │PromotionSvc   │          │ ReferralSvc   │
│               │          │               │          │               │
│ 依赖:         │          │ 依赖:         │          │ 依赖:         │
│ - nodemailer  │          │ - Redis       │          │ - Redis       │
│ - Config      │          │ - Logger      │          │ - Config      │
│ - Logger      │          │               │          │ - Logger      │
│               │          │ 无服务依赖 ✅  │          │ 无服务依赖 ✅  │
└───────────────┘          └───────────────┘          └───────────────┘
        │                           │                           │
        │                           │                           │
        ▼                           ▼                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         💵 用户余额/认证模块                             │
│                           UserService                                   │
│                                                                         │
│  依赖: Redis, Config, Logger, bcrypt, node-cache                        │
│  功能: 余额管理, 充值记录, 本地认证, 密码管理                              │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          💰 支付系统模块                                 │
│                          PaymentService                                 │
│                                                                         │
│  ┌─────────────┐    ┌──────────────┐                                   │
│  │ ZpayService │    │StripeService │  ← 支付渠道（可选择性整合）          │
│  └─────────────┘    └──────────────┘                                   │
│                                                                         │
│  依赖: UserService, PromotionService, ReferralService                   │
│  功能: 订单管理, 支付回调, 自动充值                                       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 基础依赖层（必须先整合）

### 3.1 Redis 模型适配

**原项目文件**: `src/models/redis.js`

**核心方法（需适配或实现）**:

```javascript
// 必须实现的方法
redis.getClient()        // 获取 Redis 客户端
redis.getClientSafe()    // 获取 Redis 客户端（安全版本）
redis.get(key)           // 获取值
redis.set(key, value)    // 设置值
redis.setex(key, ttl, value)  // 设置带过期时间的值
redis.del(key)           // 删除键

// 如果您的项目已有 Redis 封装，只需确保以上方法可用
```

**适配方案**:

```javascript
// 方案A: 直接复用原项目 redis.js
// 复制 src/models/redis.js 到您的项目

// 方案B: 创建适配器包装您现有的 Redis 客户端
class RedisAdapter {
  constructor(existingClient) {
    this.client = existingClient;
  }

  getClient() { return this.client; }
  getClientSafe() { return this.client; }

  async get(key) { return this.client.get(key); }
  async set(key, value) { return this.client.set(key, value); }
  async setex(key, ttl, value) { return this.client.setex(key, ttl, value); }
  async del(key) { return this.client.del(key); }
}
```

### 3.2 Logger 适配

**原项目文件**: `src/utils/logger.js`

**核心方法**:

```javascript
logger.info(message, meta)
logger.warn(message, meta)
logger.error(message, error)
logger.debug(message, meta)
logger.success(message, meta)  // 可选，可映射到 info
```

**适配方案**:

```javascript
// 方案A: 复用原项目 logger
// 复制 src/utils/logger.js

// 方案B: 映射到您现有的日志系统
const logger = {
  info: (msg, meta) => console.log('[INFO]', msg, meta),
  warn: (msg, meta) => console.warn('[WARN]', msg, meta),
  error: (msg, err) => console.error('[ERROR]', msg, err),
  debug: (msg, meta) => console.debug('[DEBUG]', msg, meta),
  success: (msg, meta) => console.log('[SUCCESS]', msg, meta),
};
```

### 3.3 Config 适配

每个功能模块需要的配置项不同，在各模块详解中说明。

---

## 4. 功能模块详解

### 4.1 📧 邮件服务 (EmailService)

**独立性**: ⭐⭐⭐⭐⭐ 完全独立，最容易整合

**文件清单**:
```
src/services/emailService.js     # 核心服务 (~330行)
```

**依赖**:
- `nodemailer` (npm 包)
- `config.email.*` (配置)
- `logger` (日志)

**所需配置**:
```javascript
config.email = {
  enabled: true,
  smtp: {
    host: 'smtp.example.com',
    port: 465,
    secure: true,
    auth: { user: 'xxx', pass: 'xxx' }
  },
  from: { name: '服务名', address: 'noreply@example.com' },
  baseUrl: 'https://your-domain.com',
  tokenTTL: {
    passwordReset: 3600,      // 1小时
    emailVerification: 86400  // 24小时
  },
  features: {
    allowPasswordReset: true,
    requireEmailVerification: false
  }
}
```

**整合步骤**:
1. 复制 `emailService.js`
2. 安装依赖: `npm install nodemailer`
3. 添加配置项
4. 修改 require 路径

**提供的功能**:
- `sendEmail(options)` - 发送邮件
- `sendPasswordResetEmail(email, token, username)` - 密码重置邮件
- `sendEmailVerificationEmail(email, token, username)` - 邮箱验证邮件

---

### 4.2 🎁 活动优惠 (PromotionService)

**独立性**: ⭐⭐⭐⭐ 基本独立

**文件清单**:
```
src/services/promotionService.js  # 核心服务 (~580行)
```

**依赖**:
- `redis` (数据存储)
- `logger` (日志)
- `crypto` (Node.js 内置)

**⚠️ 无其他服务依赖，可独立运行**

**所需配置** (通过环境变量):
```bash
PROMOTION_ENABLED=true
PROMOTION_TIER1_BONUS=30    # 24小时内 30%
PROMOTION_TIER2_BONUS=20    # 36小时内 20%
PROMOTION_TIER3_BONUS=10    # 48小时内 10%
PROMOTION_TIER4_BONUS=5     # 72小时内 5%
PROMOTION_MIN_AMOUNT=0
PROMOTION_DURATION_HOURS=72
```

**Redis Key 结构**:
```
user_promotion:{userId}      # 用户活动状态 (String/JSON)
promotion_stats:*            # 活动统计
promotion_record:{recordId}  # 活动记录
```

**整合步骤**:
1. 复制 `promotionService.js`
2. 适配 redis 和 logger 引用
3. 配置环境变量

**提供的功能**:
- `initUserPromotion(userId, username)` - 初始化用户活动
- `getUserPromotion(userId)` - 获取用户活动状态
- `applyPromotion(userId, amount)` - 评估并应用优惠
- `markPromotionUsed(userId, tier, amount, bonus)` - 标记活动已使用
- `getPromotionStats()` - 获取活动统计

**调用示例**:
```javascript
// 用户注册时初始化活动
await promotionService.initUserPromotion(userId, username);

// 用户充值时评估优惠
const result = await promotionService.applyPromotion(userId, 100);
// result: { amount: 100, bonus: 30, total: 130, applied: true, tier: 0 }

// 如果 applied=true，执行充值后标记活动已使用
if (result.applied) {
  await promotionService.markPromotionUsed(userId, result.tier, amount, result.bonus);
}
```

---

### 4.3 👥 邀请返利 (ReferralService)

**独立性**: ⭐⭐⭐⭐ 基本独立

**文件清单**:
```
src/services/referralService.js  # 核心服务 (~540行)
```

**依赖**:
- `redis` (数据存储)
- `config.referralProgram.*` (配置)
- `config.payment.exchangeRate` (汇率，可选)
- `logger` (日志)
- `crypto` (Node.js 内置)

**⚠️ 无其他服务依赖，可独立运行**

**所需配置**:
```javascript
config.referralProgram = {
  enabled: true,
  linkBaseUrl: 'https://your-domain.com/register',
  rewardUsd: 10,                    // 返利金额
  qualifiedRechargeCny: 20,         // 达标金额（人民币）
  qualifiedRechargeTypes: ['payment'], // 仅在线支付计入
  maxRewardsPerInvitee: 1,
  maxInviteesPerUser: 0,            // 0=不限
  codeLength: 8,
  rulesDescription: '返利规则说明...'
}

config.payment = {
  exchangeRate: 7.2  // 汇率（用于 CNY/USD 换算）
}
```

**Redis Key 结构**:
```
referral:code:{code}           # 邀请码 → 用户ID
referral:user:{userId}:code    # 用户ID → 邀请码
referral:invite:{inviteeId}    # 邀请记录
referral:user:{userId}:invitees # 用户的被邀请人列表
referral:stats:{userId}        # 用户邀请统计
```

**整合步骤**:
1. 复制 `referralService.js`
2. 适配 redis, config, logger 引用
3. 添加配置项

**提供的功能**:
- `getOrCreateReferralCode(userId)` - 获取/创建邀请码
- `findUserIdByCode(code)` - 通过邀请码查找用户
- `recordInvitation({inviteeId, inviteeUsername, referrerId, ...})` - 记录邀请关系
- `processInviteeRecharge({inviteeId, totalRechargeUsd, rechargeAmountUsd, recordType})` - 处理被邀请人充值
- `markRewardIssued(inviteeId, rewardAmountUsd)` - 标记返利已发放
- `getReferralInfo(userId)` - 获取邀请信息
- `listInvitees(userId, options)` - 获取被邀请人列表

**调用示例**:
```javascript
// 用户注册时，检查是否有邀请码
const inviterCode = req.query.inviter;
if (inviterCode) {
  const referrerId = await referralService.findUserIdByCode(inviterCode);
  if (referrerId) {
    await referralService.recordInvitation({
      inviteeId: newUserId,
      inviteeUsername: username,
      referrerId,
      referrerUsername: referrerUsername
    });
  }
}

// 用户充值后，检查是否触发返利
const reward = await referralService.processInviteeRecharge({
  inviteeId: userId,
  totalRechargeUsd: user.totalRecharge,
  rechargeAmountUsd: amount,
  recordType: 'payment'  // 仅在线支付
});

if (reward) {
  // 给邀请人发放返利
  await userService.rechargeBalance(reward.referrerId, reward.rewardAmountUsd, ...);
  await referralService.markRewardIssued(reward.inviteeId, reward.rewardAmountUsd);
}
```

---

### 4.4 💵 用户余额系统

**独立性**: ⭐⭐⭐ 需要适配

**文件清单**:
```
src/services/userService.js  # 核心服务（余额相关部分 ~1200行）
```

**依赖**:
- `redis` (数据存储)
- `bcrypt` (密码哈希)
- `node-cache` (内存缓存)
- `config` (配置)
- `logger` (日志)
- `crypto` (Node.js 内置)

**核心功能方法（建议提取）**:

```javascript
// 余额管理
async getBalanceInfo(userId)                    // 获取余额信息
async rechargeBalance(userId, amount, operator, remark, options)  // 充值
async deductBalance(userId, amount, operator, remark, options)    // 扣减
async getRechargeRecords(userId, options)       // 获取充值记录
async getAllRechargeRecords(options)            // 获取所有充值记录

// 用户数据结构新增字段
{
  balance: 0,           // 账户余额（美元）
  totalRecharge: 0,     // 累计充值（美元）
  lastRechargeAt: null  // 最后充值时间
}
```

**Redis Key 结构**:
```
user:{userId}                    # 用户数据（包含余额字段）
recharge_record:{recordId}       # 充值记录
recharge_records:user:{userId}   # 用户充值记录列表
recharge_records:all             # 全局充值记录列表
```

**整合方案**:

**方案A: 直接整合到现有 UserService**
```javascript
// 在您现有的 UserService 中添加以下方法

class UserService {
  // ... 您现有的方法 ...

  // 新增余额相关方法
  async getBalanceInfo(userId) { /* 从原项目复制 */ }
  async rechargeBalance(userId, amount, operator, remark, options) { /* 从原项目复制 */ }
  async deductBalance(userId, amount, operator, remark, options) { /* 从原项目复制 */ }
  async getRechargeRecords(userId, options) { /* 从原项目复制 */ }
}
```

**方案B: 创建独立 BalanceService**
```javascript
// 新建 src/services/balanceService.js
// 将余额相关功能独立出来，减少与 UserService 的耦合

class BalanceService {
  async getBalance(userId) {}
  async addBalance(userId, amount, operator, remark, options) {}
  async deductBalance(userId, amount, operator, remark, options) {}
  async getRecords(userId, options) {}
}
```

---

### 4.5 💰 支付系统

**独立性**: ⭐⭐ 依赖最多，整合最复杂

**文件清单**:
```
src/services/paymentService.js   # 支付服务抽象层 (~730行)
src/services/zpayService.js      # ZPAY 支付渠道 (~250行)
src/services/stripeService.js    # Stripe 支付渠道 (~300行)
src/routes/paymentRoutes.js      # 支付路由 (~320行)
```

**依赖**:
- `userService` - 用户余额操作
- `promotionService` - 活动优惠计算
- `referralService` - 邀请返利（可选，在回调中触发）
- `zpayService` / `stripeService` - 支付渠道
- `redis`, `config`, `logger`

**所需配置**:
```javascript
config.payment = {
  enabled: true,
  baseUrl: 'https://your-domain.com',
  minAmount: 1,
  maxAmount: 10000,
  defaultCurrency: 'CNY',
  displayCurrency: 'USD',
  exchangeRate: 7.2,
  discountRate: 1,
  orderExpireMinutes: 30,
  maxOrdersPerMinute: 3,
  packages: [],  // 充值套餐（可选）

  zpay: {
    enabled: true,
    pid: 'xxx',
    key: 'xxx',
    apiUrl: 'https://zpayz.cn',
    // ... 更多配置
  },

  stripe: {
    enabled: false,
    apiKey: 'sk_xxx',
    // ... 更多配置
  }
}
```

**Redis Key 结构**:
```
payment_order:{orderId}         # 订单详情
payment_orders_user:{userId}    # 用户订单列表
payment_orders_all              # 全局订单列表
payment_order_trade:{tradeNo}   # 交易号→订单ID映射
payment_rate_limit:{userId}     # 订单速率限制
```

**整合步骤**:
1. 先整合基础依赖（用户余额、活动优惠等）
2. 复制支付相关文件
3. 适配服务依赖引用
4. 添加支付路由
5. 配置支付渠道

**支付流程**:
```
创建订单 → 调用支付渠道 → 用户支付 → 支付回调
                                         ↓
                              验证回调 → 更新订单状态
                                         ↓
                              充值余额 → 应用活动优惠
                                         ↓
                              处理邀请返利（如有）
```

---

### 4.6 🔐 本地认证增强

**文件位置**: `src/services/userService.js` 中的认证相关方法

**核心方法**:
```javascript
// 用户注册/登录
async registerLocalUser(userData)           // 本地用户注册
async authenticateLocalUser(username, password)  // 本地用户认证

// 密码管理
async _hashPassword(password)               // bcrypt 哈希
async verifyPassword(password, passwordHash) // 验证密码
async updateUserPassword(userId, oldPassword, newPassword)  // 修改密码
async resetUserPassword(userId, newPassword)  // 管理员重置密码

// 密码重置流程
async generatePasswordResetToken(email)     // 生成重置 Token
async resetPasswordWithToken(token, newPassword)  // 使用 Token 重置

// 邮箱验证
async generateEmailVerificationToken(userId)  // 生成验证 Token
async verifyEmail(token)                     // 验证邮箱
```

**依赖**:
- `bcrypt` - 密码哈希
- `emailService` - 发送重置/验证邮件（可选）
- `redis`, `config`, `logger`, `crypto`

---

## 5. 推荐整合顺序

### 5.1 最小可行整合（仅核心功能）

```
Step 1: 基础依赖适配
        └── Redis 模型 + Logger + Config
              │
Step 2: 用户余额系统
        └── 余额管理 + 充值记录
              │
Step 3: (可选) 支付系统
        └── 订单管理 + 支付回调
```

### 5.2 完整功能整合

```
Step 1: 基础依赖适配
        ├── Redis 模型适配
        ├── Logger 适配
        └── Config 结构定义
              │
Step 2: 独立模块（可并行）
        ├── 📧 邮件服务 (EmailService)
        ├── 🎁 活动优惠 (PromotionService)
        └── 👥 邀请返利 (ReferralService)
              │
Step 3: 用户余额系统
        └── 💵 余额管理 + 充值记录 + 本地认证
              │
Step 4: 支付系统
        ├── 💰 PaymentService (核心)
        ├── ZpayService (ZPAY 渠道)
        └── StripeService (Stripe 渠道)
              │
Step 5: 路由和 API
        ├── 支付路由 (paymentRoutes.js)
        └── 用户路由扩展 (余额、活动、邀请)
              │
Step 6: 前端组件（可选）
        ├── PromotionBanner.vue
        ├── UserRechargeRecords.vue
        └── 其他 UI 组件
```

---

## 6. 整合方案选择

### 6.1 方案A: 完整复制

**适用场景**: 目标项目结构相似，希望快速整合全部功能

**步骤**:
1. 复制所有相关服务文件
2. 复制路由文件
3. 修改 require 路径
4. 添加配置项
5. 注册路由

**优点**: 快速、完整
**缺点**: 可能引入不需要的代码

### 6.2 方案B: 按需整合

**适用场景**: 只需要部分功能，或项目结构差异较大

**步骤**:
1. 确定需要的功能模块
2. 仅复制相关文件
3. 提取核心方法
4. 适配到现有服务结构

**优点**: 精简、灵活
**缺点**: 需要更多适配工作

### 6.3 方案C: 微服务化

**适用场景**: 大型项目，希望功能解耦

**步骤**:
1. 将支付系统作为独立服务
2. 通过 HTTP/消息队列 与主服务通信
3. 独立部署和维护

**架构示意**:
```
┌──────────────┐     HTTP API     ┌──────────────┐
│   主服务      │ ◄──────────────► │  支付服务     │
│  (用户管理)   │                  │ (订单/回调)   │
└──────────────┘                  └──────────────┘
                                         │
                    ┌────────────────────┴────────────────────┐
                    ▼                                         ▼
              ┌──────────┐                              ┌──────────┐
              │  ZPAY    │                              │  Stripe  │
              └──────────┘                              └──────────┘
```

---

## 7. 适配改造要点

### 7.1 用户 ID 格式适配

原项目使用 UUID 格式用户 ID。如果您的项目使用其他格式（如自增 ID），需要适配：

```javascript
// 原项目
generateUserId() {
  return crypto.randomUUID();
}

// 适配为您的格式
generateUserId() {
  return yourIdGenerator();  // 使用您项目的 ID 生成逻辑
}
```

### 7.2 货币单位适配

原项目余额以美元 (USD) 为单位。如果您需要使用其他货币：

```javascript
// 方案1: 继续使用 USD，仅在显示时转换
const displayAmount = balance * exchangeRate;

// 方案2: 修改为您的基准货币
// 需要修改所有涉及金额的地方
```

### 7.3 用户数据结构适配

原项目用户数据存储在 Redis。如果您使用 MySQL/MongoDB：

```javascript
// 原项目 (Redis)
await redis.set(`user:${userId}`, JSON.stringify(userData));

// 适配为 MySQL
await User.update({ balance, totalRecharge }, { where: { id: userId } });

// 适配为 MongoDB
await User.updateOne({ _id: userId }, { $set: { balance, totalRecharge } });
```

### 7.4 充值记录存储适配

```javascript
// 原项目 (Redis List)
await redis.lpush(`recharge_records:user:${userId}`, recordId);

// 适配为关系型数据库
await RechargeRecord.create({
  id: recordId,
  userId,
  amount,
  type,
  remark,
  createdAt: new Date()
});
```

---

## 8. 接口对接说明

### 8.1 核心接口清单

#### 余额管理

| 接口 | 方法 | 说明 |
|------|------|------|
| `/users/balance` | GET | 获取用户余额 |
| `/admin/users/:userId/recharge` | POST | 管理员充值 |
| `/admin/users/:userId/deduct` | POST | 管理员扣款 |
| `/users/recharge-records` | GET | 获取充值记录 |

#### 支付相关

| 接口 | 方法 | 说明 |
|------|------|------|
| `/payment/config` | GET | 获取支付配置 |
| `/payment/orders` | POST | 创建订单 |
| `/payment/orders` | GET | 订单列表 |
| `/payment/orders/:orderId` | GET | 订单详情 |
| `/payment/webhook/zpay` | POST/GET | ZPAY 回调 |
| `/payment/webhook/stripe` | POST | Stripe 回调 |

#### 活动优惠

| 接口 | 方法 | 说明 |
|------|------|------|
| `/users/promotion/status` | GET | 获取活动状态 |

#### 邀请返利

| 接口 | 方法 | 说明 |
|------|------|------|
| `/users/referral/info` | GET | 获取邀请信息 |
| `/users/referral/invitees` | GET | 获取被邀请人列表 |

### 8.2 服务调用示例

```javascript
// 在您的业务代码中调用

// 1. 用户充值时应用活动优惠
const promotion = await promotionService.applyPromotion(userId, amount);
const finalAmount = promotion.applied ? promotion.total : amount;
await userService.rechargeBalance(userId, finalAmount, operator, remark);
if (promotion.applied) {
  await promotionService.markPromotionUsed(userId, promotion.tier, amount, promotion.bonus);
}

// 2. 支付回调处理
app.post('/payment/webhook/zpay', async (req, res) => {
  await paymentService.handleCallback('zpay', req.body);
  res.send('success');
});

// 3. 用户注册时处理邀请关系
const inviterCode = req.body.inviterCode;
if (inviterCode) {
  const referrerId = await referralService.findUserIdByCode(inviterCode);
  if (referrerId) {
    await referralService.recordInvitation({ inviteeId, referrerId, ... });
  }
}
```

---

## 附录: 快速参考

### A. npm 依赖

```bash
npm install bcrypt node-cache p-limit nodemailer stripe
```

### B. 环境变量模板

```bash
# 邮件服务
EMAIL_ENABLED=false
SMTP_HOST=
SMTP_PORT=465
SMTP_USER=
SMTP_PASS=

# 活动优惠
PROMOTION_ENABLED=false
PROMOTION_TIER1_BONUS=30

# 邀请返利
REFERRAL_PROGRAM_ENABLED=false
REFERRAL_REWARD_USD=10

# 支付
PAYMENT_ENABLED=false
ZPAY_ENABLED=false
ZPAY_PID=
ZPAY_KEY=
STRIPE_ENABLED=false
STRIPE_SECRET_KEY=
```

### C. 文件复制清单

```bash
# 核心服务
src/services/emailService.js
src/services/promotionService.js
src/services/referralService.js
src/services/paymentService.js
src/services/zpayService.js
src/services/stripeService.js

# 路由
src/routes/paymentRoutes.js

# 配置（参考）
config/config.example.js
```
