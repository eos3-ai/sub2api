# Gemini-3-Pro-Preview API 可用性测试

## 📋 测试目的

测试通过 `https://router.aitokencloud.com` 转发的 Gemini API 中 `gemini-3-pro-preview` 模型是否可用。

## 🎯 测试范围

本测试套件包含 5 个测试用例：

1. **获取模型列表** - 验证 API 连接并确认模型存在
2. **获取模型详情** - 获取模型配置信息
3. **非流式内容生成** - 测试基本文本生成功能
4. **流式内容生成** - 测试 Server-Sent Events (SSE) 流式输出
5. **Token 计数** - 验证 Token 计算功能

## 🚀 快速开始

### 1. 安装依赖

```bash
cd gemini_api_test
pip install -r requirements.txt
```

### 2. 运行测试

```bash
python test_gemini_3_pro.py
```

### 3. 查看结果

测试完成后会在控制台输出详细结果，并在当前目录生成 `test_results.json` 文件。

## 📊 测试用例详情

### 测试 1: 获取模型列表

**API 端点**: `GET /v1beta/models`

**验证内容**:
- API 认证是否成功
- 响应状态码是否为 200
- 返回的模型列表中是否包含 `gemini-3-pro-preview`

**预期结果**: 成功返回模型列表，包含目标模型

---

### 测试 2: 获取模型详情

**API 端点**: `GET /v1beta/models/gemini-3-pro-preview`

**验证内容**:
- 是否能获取模型的详细配置
- 模型的输入/输出 Token 限制
- 模型的其他元数据

**预期结果**: 成功返回模型详细信息

---

### 测试 3: 非流式内容生成

**API 端点**: `POST /v1beta/models/gemini-3-pro-preview:generateContent`

**测试内容**: 发送简单问题 "What is 2+2? Answer in one word."

**验证内容**:
- 是否能成功生成回复
- 响应中是否包含 `candidates` 和生成的文本
- 是否返回 Token 使用统计 (`usageMetadata`)

**预期结果**: 成功生成文本回复，包含用量信息

---

### 测试 4: 流式内容生成

**API 端点**: `POST /v1beta/models/gemini-3-pro-preview:streamGenerateContent?alt=sse`

**测试内容**: 请求生成 "Count from 1 to 5, each number on a new line."

**验证内容**:
- 是否能建立 SSE 连接
- 是否能接收到多个流式事件
- 每个事件是否包含有效的 JSON 数据
- 最终是否能完整接收所有内容

**预期结果**: 成功接收流式响应，收到多个内容块

---

### 测试 5: Token 计数

**API 端点**: `POST /v1beta/models/gemini-3-pro-preview:countTokens`

**测试内容**: 计算文本 "Hello, how are you today?" 的 Token 数

**验证内容**:
- 是否能成功计算 Token
- 响应中是否包含 `totalTokens` 字段
- Token 数是否合理

**预期结果**: 成功返回 Token 计数

## ✅ 成功标准

测试全部通过需满足：

- ✓ 所有 5 个测试用例状态码均为 200
- ✓ API 认证成功
- ✓ 能够正常生成内容
- ✓ 流式输出正常工作
- ✓ Token 计数功能正常

## ❌ 常见失败场景

| HTTP 状态码 | 可能原因 | 解决方案 |
|------------|---------|---------|
| 401/403 | API Key 无效或已过期 | 检查 API Key 配置 |
| 404 | 模型不存在或 API 路径错误 | 确认模型名称和 API 端点 |
| 429 | 请求速率限制或配额耗尽 | 稍后重试或增加配额 |
| 500/502/503 | 服务器错误或上游服务问题 | 检查服务状态，稍后重试 |
| Timeout | 网络超时或服务响应慢 | 增加超时时间或检查网络 |

## 🔧 配置说明

测试脚本中的关键配置：

```python
API_BASE_URL = "https://router.aitokencloud.com"  # API 基础地址
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"  # API 密钥
MODEL = "gemini-3-pro-preview"  # 测试模型
TIMEOUT = 30  # 请求超时（秒）
```

如需修改配置，请直接编辑 `test_gemini_3_pro.py` 文件中的常量。

## 📄 输出说明

### 控制台输出

测试运行时会在控制台输出彩色格式的测试进度：

- 🟢 绿色 ✓ - 测试通过
- 🔴 红色 ✗ - 测试失败
- 🔵 蓝色 ℹ - 信息提示

### JSON 结果文件

`test_results.json` 包含：

```json
{
  "test_time": "2026-01-09T23:22:30.123456",
  "api_base_url": "https://router.aitokencloud.com",
  "model": "gemini-3-pro-preview",
  "tests": [
    {
      "name": "获取模型列表",
      "success": true,
      "result": {
        "status": "success",
        "status_code": 200,
        "elapsed": 0.52,
        ...
      }
    },
    ...
  ],
  "summary": {
    "total": 5,
    "passed": 5,
    "failed": 0,
    "conclusion": "完全可用"
  }
}
```

## 🐛 故障排除

### 问题 1: 导入错误 - 缺少模块

**错误信息**: `ModuleNotFoundError: No module named 'requests'` 或 `'sseclient'`

**解决方案**:
```bash
pip install -r requirements.txt
```

### 问题 2: SSL 证书验证失败

**错误信息**: `SSLError` 或证书相关错误

**解决方案**:
```python
# 在请求中添加 verify=False（仅用于测试）
response = requests.get(url, headers=headers, verify=False)
```

### 问题 3: 连接超时

**错误信息**: `Timeout` 或 `Connection timeout`

**解决方案**:
- 检查网络连接
- 增加 `TIMEOUT` 常量的值
- 确认 API 服务是否正常运行

### 问题 4: 认证失败

**错误信息**: `401 Unauthorized` 或 `403 Forbidden`

**解决方案**:
- 检查 API Key 是否正确
- 确认 API Key 是否有效且未过期
- 验证 API Key 的权限范围

## 📝 依赖项

- Python 3.7+
- requests >= 2.31.0 - HTTP 请求库
- sseclient-py >= 1.8.0 - Server-Sent Events 客户端

## 🔗 相关链接

- [Gemini API 文档](https://ai.google.dev/gemini-api/docs)
- [API 路由器项目](https://router.aitokencloud.com)

## 📞 支持

如遇到问题，请检查：
1. API Key 配置是否正确
2. 网络连接是否正常
3. API 服务是否在线
4. 测试结果 JSON 文件中的详细错误信息
