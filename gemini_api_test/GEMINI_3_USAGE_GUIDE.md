# Gemini 3 系列模型使用指南

## 📋 概述

本文档介绍如何通过 `router.aitokencloud.com` 使用 Gemini 3 系列模型。

**API 端点**: `https://router.aitokencloud.com`
**认证方式**: Bearer Token（使用 Sub2API 密钥）

---

## 🤖 可用模型

### 1. gemini-3-pro-preview
- **类型**: 旗舰级通用模型
- **特点**: 最强性能，适合复杂推理和长文本处理
- **响应速度**: 中等（2-6 秒）
- **推理模式**: ✅ 支持
- **最佳用途**: 复杂分析、专业写作、深度推理

### 2. gemini-3-flash-preview ⚡
- **类型**: 高速通用模型
- **特点**: 响应最快，性能优异
- **响应速度**: 快速（1-3.5 秒）
- **推理模式**: ✅ 支持
- **最佳用途**: 实时对话、快速生成、高并发场景

### 3. gemini-3-pro-image-preview (Nano Banana Pro)
- **类型**: 文本+图像生成模型
- **特点**: 支持图片生成功能
- **响应速度**: 中等（文本 2-6 秒，图片 17-23 秒）
- **推理模式**: ✅ 支持
- **图片生成**: ✅ 支持
- **最佳用途**: AI 绘图、多模态应用

---

## 🚀 快速开始

### Python 示例

```python
import requests

API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-your-api-key-here"  # 替换为你的密钥
MODEL_ID = "gemini-3-flash-preview"  # 或其他模型

def chat_with_gemini(prompt):
    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:generateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{"text": prompt}]
        }],
        "generationConfig": {
            "maxOutputTokens": 8192,
            "temperature": 1.0
        }
    }

    response = requests.post(url, headers=headers, json=payload)

    if response.status_code == 200:
        data = response.json()
        text = data["candidates"][0]["content"]["parts"][0]["text"]
        return text
    else:
        raise Exception(f"API Error: {response.status_code} - {response.text}")

# 使用示例
result = chat_with_gemini("用一句话描述人工智能的未来")
print(result)
```

### cURL 示例

```bash
curl -X POST "https://router.aitokencloud.com/v1beta/models/gemini-3-flash-preview:generateContent" \
  -H "Authorization: Bearer sk-your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [{"text": "Hello, Gemini 3!"}]
    }],
    "generationConfig": {
      "maxOutputTokens": 8192
    }
  }'
```

---

## 💡 功能使用

### 1. 基础文本生成

**适用模型**: 全部

```python
payload = {
    "contents": [{
        "parts": [{"text": "写一首关于春天的诗"}]
    }],
    "generationConfig": {
        "maxOutputTokens": 2048,
        "temperature": 0.9,  # 创意性（0-2）
        "topP": 0.95,
        "topK": 40
    }
}
```

### 2. 推理模式（Thinking Mode）

**适用模型**: 全部

启用推理模式以获得更深入的思考过程：

```python
payload = {
    "contents": [{
        "parts": [{"text": "解释量子纠缠现象"}]
    }],
    "generationConfig": {
        "maxOutputTokens": 8192,
        "thinkingConfig": {
            "thinkingBudgetTokens": 8000  # 推理预算
        }
    }
}
```

**响应结构**（包含思考过程）：

```python
data = response.json()

# 提取思考过程（如果有）
thinking = data["candidates"][0]["content"]["parts"][0].get("thought", "")

# 提取最终回答
answer = data["candidates"][0]["content"]["parts"][1]["text"]

# Token 使用统计
thinking_tokens = data["usageMetadata"]["thoughtsTokenCount"]
total_tokens = data["usageMetadata"]["totalTokenCount"]

print(f"思考 Token: {thinking_tokens}")
print(f"思考过程: {thinking}")
print(f"最终答案: {answer}")
```

### 3. 图片生成 🎨

**适用模型**: 仅 `gemini-3-pro-image-preview`

使用简单的文本描述生成图片：

```python
def generate_image(prompt):
    url = f"{API_BASE_URL}/v1beta/models/gemini-3-pro-image-preview:generateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    # 直接发送文本描述即可
    payload = {
        "contents": [{
            "parts": [{"text": prompt}]
        }]
    }

    response = requests.post(url, headers=headers, json=payload, timeout=120)

    if response.status_code == 200:
        data = response.json()

        # 提取图片数据
        parts = data["candidates"][0]["content"]["parts"]
        for part in parts:
            if "inlineData" in part:
                # 获取 base64 编码的图片
                image_base64 = part["inlineData"]["data"]
                mime_type = part["inlineData"]["mimeType"]

                # 解码并保存
                import base64
                image_bytes = base64.b64decode(image_base64)

                with open("generated_image.jpg", "wb") as f:
                    f.write(image_bytes)

                print(f"图片已保存，大小: {len(image_bytes) / 1024:.2f} KB")
                return image_bytes

    raise Exception(f"生成失败: {response.status_code}")

# 使用示例
generate_image("一只可爱的橙色猫咪坐在窗台上看着外面的花园")
generate_image("A serene mountain landscape at sunset with a lake reflecting the sky")
generate_image("未来城市天际线，有飞行汽车和霓虹灯")
```

**图片生成要点**：
- ✅ 使用 `:generateContent` 端点（不是 `:generateImages`）
- ✅ 直接发送文本描述即可
- ✅ 支持中英文提示词
- ✅ 返回 base64 编码的 JPEG 图片
- ⏱️ 生成时间约 17-23 秒
- 📦 图片大小约 1MB

### 4. 流式响应

**适用模型**: 全部

使用流式 API 获得实时响应：

```python
def stream_chat(prompt):
    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:streamGenerateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{"text": prompt}]
        }]
    }

    response = requests.post(url, headers=headers, json=payload, stream=True)

    for line in response.iter_lines():
        if line:
            # 处理 SSE 格式
            if line.startswith(b'data: '):
                import json
                data = json.loads(line[6:])

                # 提取文本片段
                if "candidates" in data:
                    text = data["candidates"][0]["content"]["parts"][0].get("text", "")
                    print(text, end="", flush=True)

    print()  # 换行

# 使用示例
stream_chat("讲一个关于AI的故事")
```

### 5. Token 计数

在发送请求前估算 token 使用量：

```python
def count_tokens(prompt):
    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:countTokens"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{"text": prompt}]
        }]
    }

    response = requests.post(url, headers=headers, json=payload)

    if response.status_code == 200:
        data = response.json()
        return data["totalTokens"]

    return None

# 使用示例
tokens = count_tokens("这是一段测试文本")
print(f"Token 数量: {tokens}")
```

---

## 📊 性能对比

基于实测数据（2026-01-10）：

| 模型 | 基本生成 | 推理测试 | 创意生成 | 速度测试 | 图片生成 |
|------|---------|---------|---------|---------|---------|
| **gemini-3-pro-preview** | 3.43s | 5.88s | 5.02s | 2.13s | ❌ |
| **gemini-3-flash-preview** ⚡ | 2.02s | 3.42s | 2.75s | 1.01s | ❌ |
| **gemini-3-pro-image-preview** | 3.69s | 6.38s | 4.85s | 2.07s | ✅ 17-23s |

### Token 使用（基础测试）

| 模型 | 输入 Token | 思考 Token | 输出 Token | 总 Token |
|------|-----------|-----------|-----------|---------|
| gemini-3-pro-preview | 10 | 197 | 0 | 207 |
| gemini-3-flash-preview | 10 | 188 | 8 | 206 |
| gemini-3-pro-image-preview | 10 | 197 | 0 | 207 |

**推理模式下的思考 Token**：
- gemini-3-pro-preview: 497 tokens
- gemini-3-flash-preview: 370 tokens
- gemini-3-pro-image-preview: 497 tokens

**推荐选择**：
- 🏃 **高速场景** → gemini-3-flash-preview
- 🧠 **复杂推理** → gemini-3-pro-preview
- 🎨 **图片生成** → gemini-3-pro-image-preview

---

## ⚙️ 配置参数

### generationConfig 完整参数

```python
"generationConfig": {
    # 输出控制
    "maxOutputTokens": 8192,        # 最大输出 token（1-8192）
    "stopSequences": ["END"],        # 停止序列

    # 采样参数
    "temperature": 1.0,              # 随机性（0-2，默认1.0）
    "topP": 0.95,                    # 核采样（0-1）
    "topK": 40,                      # Top-K 采样
    "presencePenalty": 0.0,          # 存在惩罚（-2 到 2）
    "frequencyPenalty": 0.0,         # 频率惩罚（-2 到 2）

    # 推理模式
    "thinkingConfig": {
        "thinkingBudgetTokens": 8000  # 推理预算（仅推理模式）
    },

    # 响应格式（仅图片生成模型）
    "responseModalities": ["TEXT"],   # 或 ["IMAGE"]
    "responseMimeType": "text/plain"  # 或 "application/json"
}
```

### 参数说明

**temperature**（温度）:
- `0.0` - 确定性，始终选择最可能的词
- `1.0` - 平衡（默认）
- `2.0` - 高创意性，更随机

**topP**（核采样）:
- 只考虑累积概率达到 topP 的词
- `0.95` 推荐用于大多数场景

**topK**:
- 只从概率最高的 K 个词中采样
- 通常设为 `40-50`

---

## 🔐 认证配置

### 环境变量方式

```bash
export GEMINI_API_KEY="sk-your-api-key-here"
export GEMINI_API_BASE_URL="https://router.aitokencloud.com"
```

### 代码中配置

```python
import os

API_KEY = os.environ.get("GEMINI_API_KEY", "sk-default-key")
API_BASE_URL = os.environ.get("GEMINI_API_BASE_URL", "https://router.aitokencloud.com")
```

---

## 🚨 错误处理

### 常见错误码

| 错误码 | 说明 | 解决方法 |
|-------|------|---------|
| **401** | 认证失败 | 检查 API Key 是否正确 |
| **404** | 模型不存在 | 检查模型 ID 拼写 |
| **429** | 请求过多 | 降低请求频率，添加重试逻辑 |
| **500** | 服务器错误 | 稍后重试 |
| **503** | 服务不可用 | 检查账号配置或稍后重试 |

### 错误处理示例

```python
import time

def chat_with_retry(prompt, max_retries=3):
    for attempt in range(max_retries):
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=30)

            if response.status_code == 200:
                return response.json()

            elif response.status_code == 429:
                # 速率限制，等待后重试
                wait_time = 2 ** attempt  # 指数退避
                print(f"速率限制，等待 {wait_time} 秒...")
                time.sleep(wait_time)
                continue

            elif response.status_code in [500, 503]:
                # 服务器错误，重试
                if attempt < max_retries - 1:
                    time.sleep(2)
                    continue

            else:
                # 其他错误，直接抛出
                raise Exception(f"API Error: {response.status_code} - {response.text}")

        except requests.exceptions.Timeout:
            if attempt < max_retries - 1:
                print(f"请求超时，重试 {attempt + 1}/{max_retries}...")
                time.sleep(2)
                continue
            raise

    raise Exception("达到最大重试次数")
```

---

## 💰 最佳实践

### 1. 模型选择建议

```python
# 场景 1: 实时聊天机器人
MODEL_ID = "gemini-3-flash-preview"  # 最快响应

# 场景 2: 内容创作
MODEL_ID = "gemini-3-pro-preview"
config = {
    "temperature": 1.2,  # 提高创意性
    "topP": 0.95
}

# 场景 3: 数据分析
MODEL_ID = "gemini-3-pro-preview"
config = {
    "temperature": 0.2,  # 降低随机性
    "thinkingConfig": {"thinkingBudgetTokens": 8000}  # 启用推理
}

# 场景 4: AI 绘图
MODEL_ID = "gemini-3-pro-image-preview"
# 直接发送图像描述文本即可
```

### 2. Token 优化

```python
# 使用 countTokens 预估成本
tokens = count_tokens(prompt)
if tokens > 1000:
    print(f"警告: 输入过长 ({tokens} tokens)")

# 限制输出长度
config = {
    "maxOutputTokens": 1024  # 根据需求调整
}
```

### 3. 并发请求

```python
from concurrent.futures import ThreadPoolExecutor
import requests

def process_batch(prompts):
    def call_api(prompt):
        # ... API 调用逻辑
        return response

    with ThreadPoolExecutor(max_workers=5) as executor:
        results = list(executor.map(call_api, prompts))

    return results

# 批量处理
prompts = ["问题1", "问题2", "问题3", ...]
results = process_batch(prompts)
```

### 4. 缓存策略

```python
from functools import lru_cache
import hashlib

# 简单的内存缓存
@lru_cache(maxsize=100)
def cached_chat(prompt_hash):
    # 实际 API 调用
    return chat_with_gemini(prompt)

def chat_with_cache(prompt):
    # 生成提示词的哈希
    prompt_hash = hashlib.md5(prompt.encode()).hexdigest()
    return cached_chat(prompt_hash)
```

---

## 🧪 测试工具

项目提供了完整的测试工具集：

### 快速诊断

```bash
cd gemini_api_test
python diagnose_models.py
```

输出示例：
```
检查模型: gemini-3-pro-preview
  ✓ 模型可用 (HTTP 200)

检查模型: gemini-3-flash-preview
  ✓ 模型可用 (HTTP 200)

检查模型: gemini-3-pro-image-preview
  ✓ 模型可用 (HTTP 200)

总模型数: 3
可用: 3
不可用: 0

✓ 所有模型都可用！
```

### 完整功能测试

```bash
python test_gemini_3_series.py
```

测试项目：
- ✅ 基本文本生成
- ✅ 推理模式
- ✅ 创意生成
- ✅ 速度测试

### 图片生成测试

```bash
python test_image_via_sub2api.py
```

生成的图片保存在 `generated_images/` 目录。

---

## 📚 API 参考

### 端点列表

| 端点 | 方法 | 说明 | 适用模型 |
|------|------|------|---------|
| `/v1beta/models/{model}:generateContent` | POST | 生成内容（文本或图片） | 全部 |
| `/v1beta/models/{model}:streamGenerateContent` | POST | 流式生成 | 全部 |
| `/v1beta/models/{model}:countTokens` | POST | 计算 Token | 全部 |

### 完整请求示例

```bash
POST https://router.aitokencloud.com/v1beta/models/gemini-3-flash-preview:generateContent

Headers:
  Authorization: Bearer sk-your-api-key-here
  Content-Type: application/json

Body:
{
  "contents": [{
    "parts": [{"text": "你好，Gemini 3！"}]
  }],
  "generationConfig": {
    "maxOutputTokens": 8192,
    "temperature": 1.0,
    "topP": 0.95,
    "topK": 40
  }
}
```

### 完整响应示例

```json
{
  "candidates": [{
    "content": {
      "parts": [{
        "text": "你好！我是 Gemini 3，很高兴为你服务。有什么我可以帮助你的吗？"
      }],
      "role": "model"
    },
    "finishReason": "STOP",
    "safetyRatings": [...]
  }],
  "usageMetadata": {
    "promptTokenCount": 5,
    "candidatesTokenCount": 18,
    "totalTokenCount": 23
  }
}
```

---

## 🔗 相关资源

- **Sub2API 项目**: `/Users/leizhao/Projects/claude_code_router/sub2api/`
- **测试工具**: `gemini_api_test/` 目录
- **配置文档**: `CONFIGURATION_COMPLETE.md`
- **测试报告**: `gemini_3_series_comparison.json`

---

## ❓ 常见问题

### Q: 如何选择合适的模型？

**A**:
- 需要最快响应 → `gemini-3-flash-preview`
- 需要深度推理 → `gemini-3-pro-preview`
- 需要生成图片 → `gemini-3-pro-image-preview`

### Q: 推理模式如何使用？

**A**: 在 `generationConfig` 中添加：
```python
"thinkingConfig": {
    "thinkingBudgetTokens": 8000
}
```

响应中会包含 `thought` 字段，展示模型的思考过程。

### Q: 图片生成为什么失败？

**A**: 确保：
1. 使用 `gemini-3-pro-image-preview` 模型
2. 使用 `:generateContent` 端点（不是 `:generateImages`）
3. 直接发送文本描述，不需要特殊格式
4. 设置足够的超时时间（120 秒）

### Q: 如何处理 503 错误？

**A**: 503 通常表示账号配置问题：
1. 检查账号的 `model_mapping` 是否为 `null`
2. 运行 `diagnose_models.py` 诊断
3. 联系管理员检查账号配置

### Q: Token 如何计费？

**A**: 使用 `usageMetadata` 中的统计：
- `promptTokenCount` - 输入 Token
- `candidatesTokenCount` - 输出 Token
- `thoughtsTokenCount` - 推理 Token（仅推理模式）
- `totalTokenCount` - 总计

---

## 📝 更新日志

- **2026-01-10**: 首次发布，包含全部 3 个 Gemini 3.x 模型
- **2026-01-10**: 添加图片生成功能说明
- **2026-01-10**: 完成性能测试和对比

---

*本文档基于实际测试结果编写，测试时间: 2026-01-10*
