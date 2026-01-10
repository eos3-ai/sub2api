# Sub2API Gemini 账号配置指南

## 问题诊断

### 错误信息
```
503 - No available Gemini accounts: no available accounts
```

### 原因
虽然模型在模型列表中存在，但转发服务没有为这些模型配置可用的上游 Gemini 账号。

---

## 解决步骤

### 步骤 1: 检查现有账号配置

在 sub2api 后台查看当前的 Gemini 账号配置：

1. 访问管理后台：`http://localhost:8080` （或您的部署地址）
2. 导航到 **账号管理** 页面
3. 查找 **Gemini 分组** 下的账号列表
4. 检查每个账号支持的模型列表

### 步骤 2: 添加新的 Gemini 账号

#### 方式 A: 通过 Web UI（推荐）

1. **进入账号管理页面**
   - 点击"添加账号"按钮

2. **选择账号类型**
   - 平台：选择 `Gemini`
   - 账号类型：选择 `API Key`

3. **填写账号信息**
   ```
   账号名称: Gemini 3 系列专用账号
   API Key: AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk
   Base URL: https://generativelanguage.googleapis.com （默认）
   ```

4. **配置支持的模型**
   - 在"支持的模型"列表中，勾选：
     - ✅ gemini-3-pro-preview
     - ✅ gemini-3-flash-preview
     - ✅ gemini-3-pro-image-preview

   或者选择 **"支持所有模型"**

5. **设置并发限制**（可选）
   ```
   账号并发数: 5-10（根据您的 API 配额）
   ```

6. **保存并启用**
   - 点击"保存"
   - 确保账号状态为"已启用"

#### 方式 B: 通过 API

使用 curl 命令添加账号（需要管理员权限）：

```bash
# 添加 Gemini API Key 账号
curl -X POST "http://localhost:8080/api/admin/accounts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Gemini 3 Series Account",
    "platform": "gemini",
    "type": "api_key",
    "credentials": {
      "api_key": "AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk",
      "base_url": "https://generativelanguage.googleapis.com"
    },
    "supported_models": [
      "gemini-3-pro-preview",
      "gemini-3-flash-preview",
      "gemini-3-pro-image-preview"
    ],
    "concurrency_limit": 10,
    "enabled": true
  }'
```

#### 方式 C: 通过数据库（高级）

如果直接访问数据库：

```sql
-- 插入新的 Gemini 账号
INSERT INTO accounts (
  name,
  platform,
  type,
  api_key,
  base_url,
  supported_models,
  concurrency_limit,
  enabled,
  created_at
) VALUES (
  'Gemini 3 Series Account',
  'gemini',
  'api_key',
  'AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk',
  'https://generativelanguage.googleapis.com',
  '["gemini-3-pro-preview","gemini-3-flash-preview","gemini-3-pro-image-preview"]',
  10,
  true,
  NOW()
);
```

### 步骤 3: 验证配置

添加账号后，使用测试脚本验证：

```bash
# 重新运行测试
cd gemini_api_test
python test_gemini_3_series.py
```

预期结果：所有 3 个模型都应该返回 200 状态码。

---

## 关键配置说明

### 账号类型

Sub2API 支持两种 Gemini 账号类型：

1. **API Key 类型**（推荐）
   - 使用 Google AI Studio 生成的 API Key
   - 配置简单，直接可用
   - 适合个人和小团队
   - API Key 格式：`AIza...`

2. **OAuth 类型**（企业级）
   - 使用 Google Cloud Project 的 OAuth 凭据
   - 需要配置 OAuth 2.0 客户端
   - 支持自动刷新 Access Token
   - 适合大规模部署

### 模型支持配置

账号可以配置为：

1. **支持特定模型**
   ```json
   "supported_models": [
     "gemini-3-pro-preview",
     "gemini-3-flash-preview"
   ]
   ```

2. **支持所有模型**（通配符）
   ```json
   "supported_models": ["*"]
   ```

3. **使用模型前缀匹配**
   ```json
   "supported_models": [
     "gemini-3-*",
     "gemini-2.5-*"
   ]
   ```

### 负载均衡

如果同一个模型配置了多个账号，sub2api 会自动进行负载均衡：

- **粘性会话**：同一用户的请求倾向于使用同一账号
- **负载感知**：优先选择负载较低的账号
- **故障转移**：账号失败时自动切换到其他可用账号

---

## 常见问题排查

### 问题 1: 添加账号后仍然 503

**可能原因**：
- 账号未启用（enabled = false）
- API Key 无效或已过期
- 账号并发槽位已满

**解决方案**：
```bash
# 检查账号状态
curl "http://localhost:8080/api/admin/accounts" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# 测试 API Key 是否有效
curl "https://generativelanguage.googleapis.com/v1beta/models" \
  -H "X-goog-api-key: AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk"
```

### 问题 2: 某些模型可用，某些不可用

**可能原因**：
- 账号的 `supported_models` 配置不完整
- 不同账号支持不同的模型子集

**解决方案**：
- 检查账号配置，确保包含所有需要的模型
- 或者设置 `supported_models: ["*"]` 支持所有模型

### 问题 3: API Key 配额限制

**可能原因**：
- Google API Key 有每日请求限制
- 免费版 API Key 限制更严格

**解决方案**：
- 添加多个 API Key 账号进行负载均衡
- 升级到付费的 Google Cloud Project
- 配置账号的 `rate_limit` 参数

---

## 配置示例

### 示例 1: 单个通用账号

```json
{
  "name": "Gemini Universal Account",
  "platform": "gemini",
  "type": "api_key",
  "credentials": {
    "api_key": "AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk"
  },
  "supported_models": ["*"],
  "concurrency_limit": 10,
  "enabled": true
}
```

### 示例 2: 按系列分配账号

```json
[
  {
    "name": "Gemini 3 Series",
    "supported_models": ["gemini-3-*"],
    "credentials": {"api_key": "AIza...Key1"}
  },
  {
    "name": "Gemini 2 Series",
    "supported_models": ["gemini-2.*-*"],
    "credentials": {"api_key": "AIza...Key2"}
  }
]
```

### 示例 3: 多账号负载均衡

```json
[
  {
    "name": "Gemini Account 1",
    "supported_models": ["*"],
    "credentials": {"api_key": "AIza...Key1"},
    "concurrency_limit": 5
  },
  {
    "name": "Gemini Account 2",
    "supported_models": ["*"],
    "credentials": {"api_key": "AIza...Key2"},
    "concurrency_limit": 5
  },
  {
    "name": "Gemini Account 3",
    "supported_models": ["*"],
    "credentials": {"api_key": "AIza...Key3"},
    "concurrency_limit": 5
  }
]
```

---

## 验证清单

配置完成后，检查以下项目：

- [ ] 账号已添加到系统中
- [ ] 账号状态为"已启用"
- [ ] `supported_models` 包含目标模型
- [ ] API Key 有效且未过期
- [ ] 账号并发限制合理设置
- [ ] 测试所有目标模型都返回 200
- [ ] 响应内容正确生成

---

## 相关文件位置

根据 sub2api 项目结构：

- **账号管理逻辑**: `/backend/internal/service/account_service.go`
- **模型匹配**: `/backend/internal/service/gateway_service.go`
- **账号选择**: `/backend/internal/service/gateway_service.go` 中的 `SelectAccountWithLoadAwareness()`
- **数据库表**: `accounts` 表

---

## 下一步

1. 按照上述步骤添加账号
2. 运行测试验证：`python test_gemini_3_series.py`
3. 如果仍有问题，检查 sub2api 后端日志：
   ```bash
   # 查看日志
   tail -f /path/to/sub2api/logs/app.log

   # 或者如果使用 Docker
   docker logs sub2api-backend
   ```

4. 确认账号分配策略是否符合预期
