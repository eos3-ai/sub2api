# 错误来源分析

## ❓ 问题
错误信息：`No available Gemini accounts: no available accounts`

这个错误是 **sub2api** 报出来的，还是 **上游 Google API** 报出来的？

---

## ✅ 答案：**Sub2API 报出来的**

这个错误是 **sub2api 转发服务内部**产生的，不是上游 Google API 的错误。

---

## 🔍 证据链

### 1. Handler 层（错误包装）

**文件**: `/backend/internal/handler/gemini_v1beta_handler.go`

**位置**: 第 217-221 行

```go
selection, err := h.gatewayService.SelectAccountWithLoadAwareness(
    c.Request.Context(),
    apiKey.GroupID,
    sessionKey,
    modelName,
    failedAccountIDs
)
if err != nil {
    if len(failedAccountIDs) == 0 {
        // 🔴 在这里包装错误并返回
        googleError(c, http.StatusServiceUnavailable, "No available Gemini accounts: "+err.Error())
        return
    }
    handleGeminiFailoverExhausted(c, lastFailoverStatus)
    return
}
```

### 2. 错误格式化函数

**文件**: `/backend/internal/handler/gemini_v1beta_handler.go`

**位置**: `googleError` 函数

```go
func googleError(c *gin.Context, status int, message string) {
    c.JSON(status, gin.H{
        "error": gin.H{
            "code":    status,
            "message": message,  // 完整消息：No available Gemini accounts: no available accounts
            "status":  googleapi.HTTPStatusToGoogleStatus(status),  // 转换为 "INTERNAL"
        },
    })
}
```

**输出格式**（模拟 Google API 格式）:
```json
{
  "error": {
    "code": 503,
    "message": "No available Gemini accounts: no available accounts",
    "status": "INTERNAL"
  }
}
```

### 3. Service 层（错误来源）

**文件**: `/backend/internal/service/gateway_service.go`

**位置**: `selectAccountForModelWithPlatform` 函数结尾

```go
// 3. 按优先级+最久未用选择（考虑模型支持）
var selected *Account
for i := range accounts {
    acc := &accounts[i]
    if _, excluded := excludedIDs[acc.ID]; excluded {
        continue
    }
    // 检查模型支持
    if requestedModel != "" && !s.isModelSupportedByAccount(acc, requestedModel) {
        continue
    }
    // ... 选择逻辑
}

// 🔴 如果没有找到可用账号，返回错误
if selected == nil {
    if requestedModel != "" {
        return nil, fmt.Errorf("no available accounts supporting model: %s", requestedModel)
    }
    return nil, errors.New("no available accounts")  // ← 错误的原始来源！
}
```

---

## 📊 完整流程

```
用户请求
  ↓
1️⃣ Gin Handler (gemini_v1beta_handler.go)
  ├─ 解析模型：gemini-3-flash-preview
  ├─ 调用账号选择服务
  ↓
2️⃣ Gateway Service (gateway_service.go)
  ├─ SelectAccountWithLoadAwareness()
  │   ↓
  ├─ SelectAccountForModelWithExclusions()
  │   ↓
  ├─ selectAccountForModelWithPlatform()
  │   ├─ 查询分组的可调度账号
  │   ├─ 筛选：enabled=true, platform=gemini
  │   ├─ 筛选：支持 gemini-3-flash-preview 模型
  │   ├─ 结果：找到 0 个可用账号
  │   └─ 返回错误："no available accounts"  ← 🔴 错误产生
  ↓
3️⃣ Handler 接收错误
  ├─ 包装错误信息
  ├─ 格式化为 Google API 格式
  └─ 返回给客户端
  ↓
4️⃣ 客户端收到
{
  "error": {
    "code": 503,
    "message": "No available Gemini accounts: no available accounts",
    "status": "INTERNAL"
  }
}
```

---

## 🆚 Sub2API vs Google API 错误对比

### Sub2API 内部错误（当前情况）

**特征**:
- HTTP 状态码：**503 Service Unavailable**
- 错误消息：包含 **"accounts"** 关键词
- 错误类型：**"INTERNAL"**
- 含义：**转发服务内部没有可用的上游账号**

**示例**:
```json
{
  "error": {
    "code": 503,
    "message": "No available Gemini accounts: no available accounts",
    "status": "INTERNAL"
  }
}
```

### Google API 的错误（如果是上游）

**常见的 Google API 错误**:

1. **API Key 无效**
```json
{
  "error": {
    "code": 401,
    "message": "API key not valid. Please pass a valid API key.",
    "status": "UNAUTHENTICATED"
  }
}
```

2. **模型不存在**
```json
{
  "error": {
    "code": 404,
    "message": "models/gemini-3-flash-preview is not found",
    "status": "NOT_FOUND"
  }
}
```

3. **配额不足**
```json
{
  "error": {
    "code": 429,
    "message": "Resource has been exhausted (e.g. check quota).",
    "status": "RESOURCE_EXHAUSTED"
  }
}
```

4. **上游服务不可用**
```json
{
  "error": {
    "code": 503,
    "message": "The service is currently unavailable.",
    "status": "UNAVAILABLE"
  }
}
```

**对比**：
- Google API 不会提到 **"accounts"**（它不知道您的转发架构）
- Google API 的 503 消息是 **"service unavailable"** 而不是 **"no accounts"**
- Google API 的错误状态通常是 **"UNAVAILABLE"** 而不是 **"INTERNAL"**

---

## 🔑 关键区别

| 维度 | Sub2API 错误 | Google API 错误 |
|------|-------------|----------------|
| **错误消息** | "No available **Gemini accounts**" | "API key invalid" / "Model not found" / "Service unavailable" |
| **包含 "accounts"** | ✅ 是 | ❌ 否 |
| **错误来源** | 账号选择逻辑 | API 认证/模型/服务 |
| **意义** | 转发层没有配置可用账号 | 上游 API 的问题 |
| **解决方式** | 配置 sub2api 账号 | 修复 API Key/模型名/等待恢复 |

---

## 💡 判断方法

**如何快速判断错误来源？**

1. **看错误消息关键词**
   - 包含 **"accounts"** → Sub2API 错误
   - 包含 **"API key"** / **"quota"** / **"model not found"** → Google API 错误

2. **看错误状态**
   - **"INTERNAL"** + 503 + "accounts" → Sub2API
   - **"UNAUTHENTICATED"** / **"NOT_FOUND"** / **"RESOURCE_EXHAUSTED"** → Google API

3. **测试方法**
   直接调用 Google API（绕过转发服务）：
   ```bash
   curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent" \
     -H "X-goog-api-key: AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk" \
     -H "Content-Type: application/json" \
     -d '{"contents":[{"parts":[{"text":"test"}]}]}'
   ```

   - 如果直接调用成功 → 问题在 Sub2API（账号配置）
   - 如果直接调用也失败 → 问题在上游 API

---

## 📌 结论

**错误 "No available Gemini accounts: no available accounts" 是 Sub2API 报出来的**

**原因**：
- Sub2API 在账号选择逻辑中找不到支持 `gemini-3-flash-preview` 和 `gemini-3-pro-image-preview` 的可用账号
- 当前只有一个账号配置了 `gemini-3-pro-preview` 的支持

**解决方案**：
1. 修改现有 Gemini 账号配置，添加对这两个模型的支持
2. 或者添加新的 Gemini 账号并配置支持这些模型
3. 或者将账号设置为"支持所有模型"（通配符）

**验证**：
配置完成后运行：
```bash
python diagnose_models.py
```

所有模型应该显示 ✓ 可用。
