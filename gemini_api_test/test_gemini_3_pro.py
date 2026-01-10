#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Gemini-3-Pro-Preview API 测试脚本

测试通过 router.aitokencloud.com 转发的 Gemini API 是否可用。
"""

import requests
import sseclient
import json
import time
from datetime import datetime
from typing import Dict, Any, Tuple

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"
MODEL = "gemini-3-pro-preview"
TIMEOUT = 30  # 请求超时时间（秒）

# 颜色输出
class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text: str):
    """打印测试标题"""
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}\n")

def print_success(text: str):
    """打印成功信息"""
    print(f"{Colors.OKGREEN}✓ {text}{Colors.ENDC}")

def print_error(text: str):
    """打印错误信息"""
    print(f"{Colors.FAIL}✗ {text}{Colors.ENDC}")

def print_info(text: str):
    """打印信息"""
    print(f"{Colors.OKCYAN}ℹ {text}{Colors.ENDC}")

def get_headers(content_type: str = "application/json") -> Dict[str, str]:
    """获取请求头"""
    return {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": content_type
    }

def test_list_models() -> Tuple[bool, Dict[str, Any]]:
    """
    测试 1: 获取模型列表
    验证 API 连接是否正常，并确认 gemini-3-pro-preview 是否在可用模型列表中
    """
    print_header("测试 1: 获取模型列表")

    url = f"{API_BASE_URL}/v1beta/models"
    start_time = time.time()

    try:
        print_info(f"请求 URL: {url}")
        response = requests.get(url, headers=get_headers(), timeout=TIMEOUT)
        elapsed = time.time() - start_time

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f}秒")

        if response.status_code == 200:
            data = response.json()
            models = data.get("models", [])
            print_info(f"返回 {len(models)} 个模型")

            # 检查是否包含目标模型
            model_names = [m.get("name", "") for m in models]
            target_found = any(MODEL in name for name in model_names)

            if target_found:
                print_success(f"找到目标模型: {MODEL}")
                return True, {
                    "status": "success",
                    "status_code": response.status_code,
                    "elapsed": elapsed,
                    "model_count": len(models),
                    "target_found": True
                }
            else:
                print_error(f"未找到目标模型: {MODEL}")
                print_info(f"可用模型: {', '.join([m.get('name', 'unknown')[:50] for m in models[:5]])}")
                return False, {
                    "status": "failed",
                    "status_code": response.status_code,
                    "elapsed": elapsed,
                    "model_count": len(models),
                    "target_found": False,
                    "error": f"模型 {MODEL} 不在列表中"
                }
        else:
            print_error(f"请求失败: {response.status_code}")
            print_error(f"响应内容: {response.text[:200]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:200]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_get_model_details() -> Tuple[bool, Dict[str, Any]]:
    """
    测试 2: 获取模型详情
    获取 gemini-3-pro-preview 的详细信息和配置
    """
    print_header("测试 2: 获取模型详情")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL}"
    start_time = time.time()

    try:
        print_info(f"请求 URL: {url}")
        response = requests.get(url, headers=get_headers(), timeout=TIMEOUT)
        elapsed = time.time() - start_time

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f}秒")

        if response.status_code == 200:
            data = response.json()
            print_success("成功获取模型详情")

            # 打印关键信息
            if "name" in data:
                print_info(f"模型名称: {data['name']}")
            if "displayName" in data:
                print_info(f"显示名称: {data['displayName']}")
            if "inputTokenLimit" in data:
                print_info(f"输入 Token 限制: {data['inputTokenLimit']}")
            if "outputTokenLimit" in data:
                print_info(f"输出 Token 限制: {data['outputTokenLimit']}")

            return True, {
                "status": "success",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "model_info": data
            }
        else:
            print_error(f"请求失败: {response.status_code}")
            print_error(f"响应内容: {response.text[:200]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:200]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_generate_content() -> Tuple[bool, Dict[str, Any]]:
    """
    测试 3: 非流式内容生成
    测试模型的基本文本生成能力
    """
    print_header("测试 3: 非流式内容生成")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL}:generateContent"
    start_time = time.time()

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "What is 2+2? Answer in one word."}]
        }],
        "generationConfig": {
            "maxOutputTokens": 50,
            "temperature": 0.7
        }
    }

    try:
        print_info(f"请求 URL: {url}")
        print_info(f"请求内容: {payload['contents'][0]['parts'][0]['text']}")

        response = requests.post(
            url,
            headers=get_headers(),
            json=payload,
            timeout=TIMEOUT
        )
        elapsed = time.time() - start_time

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f}秒")

        if response.status_code == 200:
            data = response.json()
            print_success("成功生成内容")

            # 提取生成的文本
            if "candidates" in data and len(data["candidates"]) > 0:
                candidate = data["candidates"][0]
                if "content" in candidate and "parts" in candidate["content"]:
                    generated_text = candidate["content"]["parts"][0].get("text", "")
                    print_info(f"生成的内容: {generated_text}")

                # 提取用量信息
                if "usageMetadata" in data:
                    usage = data["usageMetadata"]
                    print_info(f"Token 使用: 输入={usage.get('promptTokenCount', 0)}, 输出={usage.get('candidatesTokenCount', 0)}, 总计={usage.get('totalTokenCount', 0)}")

            return True, {
                "status": "success",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "response": data
            }
        else:
            print_error(f"请求失败: {response.status_code}")
            print_error(f"响应内容: {response.text[:200]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:200]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_stream_generate_content() -> Tuple[bool, Dict[str, Any]]:
    """
    测试 4: 流式内容生成
    测试模型的流式输出能力（Server-Sent Events）
    """
    print_header("测试 4: 流式内容生成")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL}:streamGenerateContent?alt=sse"
    start_time = time.time()

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "Count from 1 to 5, each number on a new line."}]
        }],
        "generationConfig": {
            "maxOutputTokens": 100
        }
    }

    try:
        print_info(f"请求 URL: {url}")
        print_info(f"请求内容: {payload['contents'][0]['parts'][0]['text']}")

        response = requests.post(
            url,
            headers=get_headers(),
            json=payload,
            stream=True,
            timeout=TIMEOUT
        )

        print_info(f"HTTP 状态码: {response.status_code}")

        if response.status_code == 200:
            print_success("开始接收流式响应")

            chunks = []
            event_count = 0

            try:
                client = sseclient.SSEClient(response)
                for event in client.events():
                    event_count += 1
                    if event.data:
                        try:
                            data = json.loads(event.data)
                            chunks.append(data)

                            # 提取部分文本
                            if "candidates" in data and len(data["candidates"]) > 0:
                                candidate = data["candidates"][0]
                                if "content" in candidate and "parts" in candidate["content"]:
                                    text = candidate["content"]["parts"][0].get("text", "")
                                    if text:
                                        print_info(f"接收到内容块 #{event_count}: {text[:50]}")
                        except json.JSONDecodeError:
                            pass

                elapsed = time.time() - start_time
                print_success(f"完成流式响应，共接收 {event_count} 个事件")
                print_info(f"总响应时间: {elapsed:.2f}秒")

                return True, {
                    "status": "success",
                    "status_code": response.status_code,
                    "elapsed": elapsed,
                    "event_count": event_count,
                    "chunks": len(chunks)
                }
            except Exception as stream_error:
                elapsed = time.time() - start_time
                print_error(f"流式处理异常: {str(stream_error)}")
                return False, {
                    "status": "error",
                    "elapsed": elapsed,
                    "error": f"流式处理失败: {str(stream_error)}"
                }
        else:
            elapsed = time.time() - start_time
            print_error(f"请求失败: {response.status_code}")
            print_error(f"响应内容: {response.text[:200]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:200]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_count_tokens() -> Tuple[bool, Dict[str, Any]]:
    """
    测试 5: Token 计数
    测试 Token 计数功能
    """
    print_header("测试 5: Token 计数")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL}:countTokens"
    start_time = time.time()

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "Hello, how are you today?"}]
        }]
    }

    try:
        print_info(f"请求 URL: {url}")
        print_info(f"测试文本: {payload['contents'][0]['parts'][0]['text']}")

        response = requests.post(
            url,
            headers=get_headers(),
            json=payload,
            timeout=TIMEOUT
        )
        elapsed = time.time() - start_time

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f}秒")

        if response.status_code == 200:
            data = response.json()
            print_success("成功计算 Token")

            if "totalTokens" in data:
                print_info(f"Token 总数: {data['totalTokens']}")

            return True, {
                "status": "success",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "token_count": data.get("totalTokens", 0)
            }
        else:
            print_error(f"请求失败: {response.status_code}")
            print_error(f"响应内容: {response.text[:200]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:200]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def save_results(results: Dict[str, Any]):
    """保存测试结果到 JSON 文件"""
    output_file = "test_results.json"
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        print_success(f"测试结果已保存到: {output_file}")
    except Exception as e:
        print_error(f"保存结果失败: {str(e)}")

def main():
    """主函数"""
    print(f"\n{Colors.BOLD}{'='*60}")
    print(f"Gemini-3-Pro-Preview API 可用性测试")
    print(f"{'='*60}{Colors.ENDC}\n")

    print_info(f"API Base URL: {API_BASE_URL}")
    print_info(f"目标模型: {MODEL}")
    print_info(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    # 执行所有测试
    test_functions = [
        ("获取模型列表", test_list_models),
        ("获取模型详情", test_get_model_details),
        ("非流式内容生成", test_generate_content),
        ("流式内容生成", test_stream_generate_content),
        ("Token 计数", test_count_tokens)
    ]

    results = {
        "test_time": datetime.now().isoformat(),
        "api_base_url": API_BASE_URL,
        "model": MODEL,
        "tests": []
    }

    passed = 0
    failed = 0

    for name, test_func in test_functions:
        success, result = test_func()
        results["tests"].append({
            "name": name,
            "success": success,
            "result": result
        })

        if success:
            passed += 1
        else:
            failed += 1

        # 测试之间稍作延迟
        time.sleep(1)

    # 打印总结
    print_header("测试总结")
    print_info(f"总测试数: {passed + failed}")
    print_success(f"通过: {passed}")
    if failed > 0:
        print_error(f"失败: {failed}")

    # 最终结论
    print(f"\n{Colors.BOLD}{'='*60}")
    if passed == len(test_functions):
        print(f"{Colors.OKGREEN}{Colors.BOLD}✓ 结论: gemini-3-pro-preview 完全可用{Colors.ENDC}")
    elif passed > 0:
        print(f"{Colors.WARNING}{Colors.BOLD}⚠ 结论: gemini-3-pro-preview 部分可用 ({passed}/{len(test_functions)} 测试通过){Colors.ENDC}")
    else:
        print(f"{Colors.FAIL}{Colors.BOLD}✗ 结论: gemini-3-pro-preview 不可用{Colors.ENDC}")
    print(f"{Colors.BOLD}{'='*60}{Colors.ENDC}\n")

    # 保存结果
    results["summary"] = {
        "total": passed + failed,
        "passed": passed,
        "failed": failed,
        "conclusion": "完全可用" if passed == len(test_functions) else ("部分可用" if passed > 0 else "不可用")
    }

    save_results(results)

    return 0 if passed == len(test_functions) else 1

if __name__ == "__main__":
    exit(main())
