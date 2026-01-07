# Sub2API Docker 部署配置指南 (v0.1.41 → dev 合并版本)

**文档生成时间:** 2025-01-07
**适用版本:** dev 分支 (commit 9219bb0)
**更新内容:** 122 个新提交，包含重要功能增强和 bug 修复

---

## 📋 新功能概览

本次更新包含 122 个提交，主要新增功能如下：

### 🎨 1. **图片生成计费功能** (重要)
- ✅ 支持 `gemini-3-pro-image` 模型的图片生成计费
- ✅ 按分辨率计费：1K、2K、4K 三档
- ✅ 新增数据库字段：
  - `groups.image_price_1k/2k/4k` - 分组图片定价
  - `usage_logs.image_count` - 图片生成数量
  - `usage_logs.image_size` - 图片分辨率

**影响范围:** Antigravity 平台的图片生成 API

---

### 📝 2. **账号备注功能**
- ✅ 管理员可为账号添加备注信息
- ✅ 新增字段：`accounts.notes` (TEXT)
- ✅ 前端界面支持备注编辑和显示

---

### ⏸️ 3. **临时不可调度功能**
- ✅ 支持临时禁用账号调度（不删除账号）
- ✅ 新增数据库字段：`accounts.temp_unschedulable`
- ✅ 适用场景：临时维护、调试、测试

---

### 📊 4. **用量统计增强**
- ✅ 支持 Token 统计的 M 单位显示 (Million)
- ✅ 修复跨时区日期查询不准确问题
- ✅ 优化管理员用量页面功能和体验
- ✅ 添加成本 Tooltip 明细

---

### 🔒 5. **安全性增强**
- ✅ CSP 策略优化（支持 Cloudflare Turnstile、Google Fonts）
- ✅ URL 安全配置优化，默认值改为开发友好模式
- ✅ 依赖安全更新（golang.org/x/crypto v0.46.0）
- ✅ 前端依赖漏洞修复和安全扫描强化

**重要变更:**
```bash
# .env 文件中的默认值已更改为开发友好模式
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=true
```

---

### 🛠️ 6. **网关和并发控制优化**
- ✅ 修复账号跨分组调度问题
- ✅ 修复 `cache_control` 块超限问题
- ✅ 优化 Claude Code 检测逻辑
- ✅ 修复 `wrapReleaseOnDone` goroutine 泄露问题
- ✅ 修复 Antigravity 账户刷新 token 500 错误

---

### 💰 7. **计费系统修复**
- ✅ 修复计费漏洞，允许余额透支策略
- ✅ 修复零值字段无法保存的问题
- ✅ 支持图片生成计费

---

### 🌐 8. **前端组件优化**
- ✅ 迁移到 pnpm 包管理器
- ✅ 统一 SVG 图标管理（新增 Icon 组件）
- ✅ 优化所有管理页面搜索框布局
- ✅ 修复表单校验和错误提示
- ✅ 用户管理页面添加 ID 列

---

### 🐳 9. **Docker 部署改进**
- ✅ 支持通过环境变量配置挂载的配置文件路径 (`CONFIG_FILE`)
- ✅ 更新 Redis 镜像至 8-alpine
- ✅ 优化健康检查和启动依赖
- ✅ 强调 JWT_SECRET 配置重要性（防止重启后登录失效）

---

## 🔧 Docker Compose 配置变更

### 无需修改的配置

**好消息:** 如果你已经有运行中的 Docker Compose 部署，**核心配置无需修改**。以下配置保持不变：

- ✅ PostgreSQL 连接配置
- ✅ Redis 连接配置
- ✅ 基础服务端口映射
- ✅ 健康检查配置
- ✅ 网络和卷配置

---

### 建议更新的配置

#### 1. **JWT_SECRET** (强烈推荐)

**问题:** 如果不设置固定的 JWT_SECRET，容器重启后会生成新的密钥，导致所有用户被强制登出。

**解决方案:** 在 `.env` 文件中设置固定的 JWT_SECRET

```bash
# 生成一个安全的密钥
openssl rand -hex 32

# 将生成的密钥添加到 .env 文件
JWT_SECRET=your_generated_secret_here
```

**修改位置:** `deploy/.env`

---

#### 2. **CONFIG_FILE 环境变量** (可选)

**新功能:** 现在可以自定义配置文件挂载路径

**默认行为:**
```yaml
volumes:
  - ${CONFIG_FILE:-./config.yaml}:/app/data/config.yaml:ro
```

**使用场景:** 如果你想使用自定义配置文件位置：

```bash
# 在 .env 中添加
CONFIG_FILE=/path/to/your/custom/config.yaml
```

**修改位置:** `deploy/.env`

---

#### 3. **安全配置默认值** (已更改)

**变更说明:** URL 安全配置默认值已改为开发友好模式

**新的默认值 (deploy/.env.example):**
```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=true
```

**影响:**
- ✅ 开发环境更友好（允许 HTTP、本地 IP）
- ⚠️ 生产环境建议改为 `false` 以增强安全性

**修改位置:** `deploy/.env`

---

#### 4. **Redis 镜像版本更新** (自动)

**变更:**
```yaml
# 旧版本
image: redis:7-alpine

# 新版本
image: redis:8-alpine
```

**影响:** 无需手动修改，`docker-compose pull` 会自动拉取新镜像

---

## 📝 环境变量 (.env) 配置清单

### 必须检查的配置项

#### ✅ 1. **数据库密码** (必须)
```bash
POSTGRES_PASSWORD=your_secure_password_here
```
**说明:** 如果未设置，Docker Compose 会报错退出

---

#### ✅ 2. **JWT 密钥** (强烈推荐)
```bash
# 生成方式
JWT_SECRET=$(openssl rand -hex 32)
```
**说明:** 防止容器重启后所有用户登录失效

---

#### ✅ 3. **安全配置** (根据环境调整)

**开发环境/内网:**
```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=true
```

**生产环境/公网:**
```bash
SECURITY_URL_ALLOWLIST_ENABLED=true
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=false
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=false
```

---

### 可选配置项

#### 🎨 **图片生成计费配置**

如果你使用 Antigravity 平台的 `gemini-3-pro-image` 模型，需要在**分组配置**中设置图片定价：

**配置方式:** 通过管理后台 → 分组管理 → 编辑分组

**字段说明:**
- `image_price_1k`: 1K 分辨率图片单价 (USD)
- `image_price_2k`: 2K 分辨率图片单价 (USD)
- `image_price_4k`: 4K 分辨率图片单价 (USD)

**示例定价:**
```yaml
groups:
  - name: "默认分组"
    image_price_1k: 0.01
    image_price_2k: 0.02
    image_price_4k: 0.04
```

---

#### 💳 **支付配置** (Payment Module)

**说明:** `.env.example` 已包含完整的支付配置示例，默认禁用

```bash
# 启用支付模块
PAYMENT_ENABLED=false

# 如需启用，参考 deploy/.env.example 中的详细配置
```

**涉及功能:**
- ZPay 支付（支付宝/微信）
- Stripe 支付
- 充值套餐配置
- 首充优惠
- 邀请奖励

---

#### 🔐 **Gemini OAuth 配置**

**说明:** 如果你使用 Gemini OAuth 账号，需要配置客户端 ID 和密钥

```bash
GEMINI_OAUTH_CLIENT_ID=your_client_id
GEMINI_OAUTH_CLIENT_SECRET=your_client_secret
```

**参考文档:** [deploy/.env.example](deploy/.env.example) 第 91-125 行

---

## 🚀 部署步骤

### 首次部署

```bash
cd deploy

# 1. 复制环境变量模板
cp .env.example .env

# 2. 编辑 .env 文件，至少设置以下内容
nano .env
# - POSTGRES_PASSWORD=<设置数据库密码>
# - JWT_SECRET=<使用 openssl rand -hex 32 生成>

# 3. 启动服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f sub2api

# 5. 访问应用
# http://localhost:8080
```

---

### 已有部署升级

```bash
cd deploy

# 1. 拉取新镜像
docker-compose pull

# 2. 检查并更新 .env 配置
# - 添加 JWT_SECRET（如果未设置）
# - 检查安全配置默认值

# 3. 停止并重新创建容器
docker-compose down
docker-compose up -d

# 4. 查看启动日志
docker-compose logs -f sub2api

# 5. 数据库迁移（自动执行）
# 应用启动时会自动运行新的迁移脚本：
# - 028_add_account_notes.sql
# - 028_group_image_pricing.sql
# - 029_usage_log_image_fields.sql
```

---

## 🗄️ 数据库迁移

### 新增迁移脚本

本次更新包含以下数据库迁移：

| 文件名 | 说明 |
|--------|------|
| `020_add_temp_unschedulable.sql` | 添加临时不可调度字段 |
| `026_ops_metrics_aggregation_tables.sql` | 运维指标聚合表 |
| `027_usage_billing_consistency.sql` | 用量计费一致性修复 |
| `028_add_account_notes.sql` | 账号备注字段 |
| `028_group_image_pricing.sql` | 分组图片定价字段 |
| `029_usage_log_image_fields.sql` | 用量日志图片统计字段 |

**执行方式:** 自动执行（应用启动时）

**验证方式:**
```sql
-- 连接到数据库
docker exec -it sub2api-postgres psql -U sub2api -d sub2api

-- 检查新字段是否存在
\d accounts
\d groups
\d usage_logs
```

---

## ⚠️ 注意事项

### 1. **JWT_SECRET 配置**

**问题:** 容器重启后所有用户登录失效

**原因:** 未设置固定的 JWT_SECRET，每次启动生成新密钥

**解决方案:**
```bash
# 在 .env 中添加
JWT_SECRET=$(openssl rand -hex 32)
```

---

### 2. **时区配置**

**影响:** 用量统计、订阅到期时间、日志时间戳

**配置方式:**
```bash
# .env 文件
TZ=Asia/Shanghai
```

**支持的时区:** 参考 [IANA Time Zone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

---

### 3. **安全配置**

**开发环境:**
```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=true
```

**生产环境:**
```bash
SECURITY_URL_ALLOWLIST_ENABLED=true
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=false
SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS=false
```

---

### 4. **Redis 密码**

**说明:** 默认为空（无密码），适合本地开发

**生产环境建议设置:**
```bash
REDIS_PASSWORD=your_redis_password
```

---

### 5. **外部数据库连接**

**场景:** 使用宿主机或外部的 PostgreSQL/Redis

**配置方式:** 创建 `docker-compose.override.yml`

```bash
# 复制示例文件
cp docker-compose.override.yml.example docker-compose.override.yml

# 编辑配置（参考文件中的 Scenario 1）
nano docker-compose.override.yml
```

**示例配置:**
```yaml
services:
  sub2api:
    depends_on: []
    extra_hosts:
      - "host.docker.internal:host-gateway"
    environment:
      DATABASE_HOST: host.docker.internal
      DATABASE_PORT: "5432"
      REDIS_HOST: host.docker.internal
      REDIS_PORT: "6379"

  postgres:
    deploy:
      replicas: 0

  redis:
    deploy:
      replicas: 0
```

---

## 🔍 故障排查

### 问题 1: 容器启动失败

**检查步骤:**
```bash
# 1. 查看日志
docker-compose logs sub2api

# 2. 检查数据库连接
docker-compose logs postgres

# 3. 检查 Redis 连接
docker-compose logs redis

# 4. 检查环境变量
docker-compose config
```

---

### 问题 2: 数据库迁移失败

**症状:** 日志中出现 `migration failed` 错误

**解决方案:**
```bash
# 1. 进入容器
docker exec -it sub2api-postgres psql -U sub2api -d sub2api

# 2. 检查迁移表
SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 10;

# 3. 手动运行迁移脚本
\i /path/to/migration.sql
```

---

### 问题 3: 所有用户登录失效

**原因:** JWT_SECRET 未设置或容器重启后改变

**解决方案:**
```bash
# 1. 在 .env 中设置固定的 JWT_SECRET
JWT_SECRET=$(openssl rand -hex 32)

# 2. 重启容器
docker-compose restart sub2api
```

---

### 问题 4: 图片生成计费不工作

**检查步骤:**
```bash
# 1. 检查数据库字段是否存在
docker exec -it sub2api-postgres psql -U sub2api -d sub2api
\d groups

# 2. 检查分组配置
SELECT id, name, image_price_1k, image_price_2k, image_price_4k FROM groups;

# 3. 设置图片定价（管理后台 → 分组管理）
```

---

## 📊 配置对比表

### docker-compose.yml 变更

| 配置项 | 旧版本 | 新版本 | 说明 |
|--------|--------|--------|------|
| Redis 镜像 | `redis:7-alpine` | `redis:8-alpine` | 版本升级 |
| 配置文件挂载 | 固定路径 | `${CONFIG_FILE:-./config.yaml}` | 支持自定义路径 |
| 安全配置 | 无默认值 | 开发友好默认值 | `.env` 中配置 |

---

### .env 文件变更

| 环境变量 | 状态 | 说明 |
|---------|------|------|
| `JWT_SECRET` | 🆕 强烈推荐 | 防止重启后登录失效 |
| `SECURITY_URL_ALLOWLIST_ENABLED` | ✏️ 默认值改为 `false` | 开发友好模式 |
| `SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP` | ✏️ 默认值改为 `true` | 允许 HTTP |
| `SECURITY_URL_ALLOWLIST_ALLOW_PRIVATE_HOSTS` | ✏️ 默认值改为 `true` | 允许本地 IP |
| `CONFIG_FILE` | 🆕 可选 | 自定义配置文件路径 |

---

## 📚 参考文档

- **环境变量完整列表:** [deploy/.env.example](deploy/.env.example)
- **Docker Compose 配置:** [deploy/docker-compose.yml](deploy/docker-compose.yml)
- **Override 示例:** [deploy/docker-compose.override.yml.example](deploy/docker-compose.override.yml.example)
- **数据库迁移脚本:** [backend/migrations/](backend/migrations/)
- **v0.1.41 合并报告:** [MERGE_v0.1.41_REPORT.md](MERGE_v0.1.41_REPORT.md)

---

## ✅ 快速检查清单

升级前请确认以下事项：

- [ ] 已备份数据库和配置文件
- [ ] 已设置 `POSTGRES_PASSWORD`
- [ ] 已设置 `JWT_SECRET`（防止重启后登录失效）
- [ ] 已检查安全配置（开发/生产环境）
- [ ] 已拉取最新镜像 (`docker-compose pull`)
- [ ] 已查看 [deploy/.env.example](deploy/.env.example) 了解新配置项
- [ ] 已准备好回滚方案（数据库备份）

---

**更新日志:** 本指南基于 dev 分支 (commit 9219bb0)，包含 122 个新提交的功能和修复。

**反馈渠道:** 如有问题，请在 GitHub Issues 中反馈。
