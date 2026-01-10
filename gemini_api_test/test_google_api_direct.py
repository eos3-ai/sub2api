#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
直接测试 Google API（绕过 sub2api）

使用代理直接访问 Google API，验证模型是否在上游可用
"""

import requests
import os
import json
from typing import Dict, Tuple

# 设置代理
os.environ['http_proxy'] = 'http://127.0.0.1:7897'
os.environ['https_proxy'] = 'http://127.0.0.1:7897'
os.environ['all_proxy'] = 'socks5://127.0.0.1:7897'

# Google API 配置
GOOGLE_API_BASE = "https://generativelanguage.googleapis.com"
GOOGLE_API_KEY = "AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk"

# 测试模型
MODELS = [
    "gemini-3-pro-preview",
    "gemini-3-flash-preview",
    "gemini-3-pro-image-preview"
]

# 颜色输出
class Colors:
    HEADER = '\033[95m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    OKCYAN = '\033[96m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text: str):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*70}")
    print(f"{text}")
    print(f"{'='*70}{Colors.ENDC}\n")

def print_success(text: str):
    print(f"{Colors.OKGREEN}✓ {text}{Colors.ENDC}")

def print_error(text: str):
    print(f"{Colors.FAIL}✗ {text}{Colors.ENDC}")

def print_info(text: str):
    print(f"{Colors.OKCYAN}ℹ {text}{Colors.ENDC}")

def print_warning(text: str):
    print(f"{Colors.WARNING}⚠ {text}{Colors.ENDC}")

def test_google_api_model(model_id: str) -> Tuple[bool, Dict]:
    """直接测试 Google API 中的模型"""
    print(f"\n{Colors.BOLD}测试模型: {model_id}{Colors.ENDC}")

    # 使用 X-goog-api-key header（推荐方式）
    url = f"{GOOGLE_API_BASE}/v1beta/models/{model_id}:generateContent"
    headers = {
        "Content-Type": "application/json",
        "X-goog-api-key": GOOGLE_API_KEY
    }

    payload = {
        "contents": [{
            "parts": [{"text": "Say hello in one word."}]
        }],
        "generationConfig": {
            "maxOutputTokens": 50
        }
    }

    print_info(f"请求 URL: {url}")
    print_info(f"使用认证: X-goog-api-key header")

    try:
        # 使用代理
        proxies = {
            'http': 'http://127.0.0.1:7897',
            'https': 'http://127.0.0.1:7897'
        }

        response = requests.post(
            url,
            headers=headers,
            json=payload,
            proxies=proxies,
            timeout=30
        )

        print_info(f"HTTP 状态码: {response.status_code}")

        result = {
            "status_code": response.status_code,
            "available": response.status_code == 200
        }

        if response.status_code == 200:
            data = response.json()
            print_success("模型在 Google API 中可用！")

            # 提取响应内容
            if "candidates" in data:
                candidates = data["candidates"]
                print_info(f"返回 {len(candidates)} 个候选响应")

                if len(candidates) > 0 and "content" in candidates[0]:
                    content = candidates[0]["content"]
                    if "parts" in content and len(content["parts"]) > 0:
                        text = content["parts"][0].get("text", "")
                        if text:
                            print_info(f"生成内容: {text}")

            # 提取用量信息
            if "usageMetadata" in data:
                usage = data["usageMetadata"]
                tokens = usage.get("totalTokenCount", 0)
                thinking = usage.get("thoughtsTokenCount", 0)
                print_info(f"Token 使用: {tokens}")
                if thinking > 0:
                    print_warning(f"推理 Token: {thinking}")

            result["response"] = data

        else:
            print_error(f"模型不可用或请求失败")

            try:
                error_data = response.json()
                error_msg = error_data.get("error", {}).get("message", "Unknown error")
                error_status = error_data.get("error", {}).get("status", "UNKNOWN")

                print_error(f"错误消息: {error_msg}")
                print_error(f"错误状态: {error_status}")

                result["error"] = {
                    "message": error_msg,
                    "status": error_status
                }

                # 分析错误类型
                if response.status_code == 404:
                    print_warning("原因: 模型不存在或路径错误")
                elif response.status_code == 401 or response.status_code == 403:
                    print_warning("原因: API Key 无效或权限不足")
                elif response.status_code == 429:
                    print_warning("原因: 配额耗尽或速率限制")
                elif response.status_code == 400:
                    print_warning("原因: 请求参数错误")

            except:
                result["error"] = {"message": response.text[:300], "status": "UNKNOWN"}
                print_error(f"响应内容: {response.text[:300]}")

        return result["available"], result

    except requests.exceptions.ProxyError as e:
        print_error(f"代理连接失败: {str(e)}")
        return False, {"error": f"Proxy error: {str(e)}"}
    except requests.exceptions.Timeout as e:
        print_error(f"请求超时: {str(e)}")
        return False, {"error": f"Timeout: {str(e)}"}
    except Exception as e:
        print_error(f"发生异常: {str(e)}")
        return False, {"error": str(e)}

def test_model_info(model_id: str) -> Tuple[bool, Dict]:
    """获取模型详细信息"""
    print(f"\n{Colors.BOLD}获取模型信息: {model_id}{Colors.ENDC}")

    url = f"{GOOGLE_API_BASE}/v1beta/models/{model_id}"
    headers = {
        "X-goog-api-key": GOOGLE_API_KEY
    }

    print_info(f"请求 URL: {url}")

    try:
        proxies = {
            'http': 'http://127.0.0.1:7897',
            'https': 'http://127.0.0.1:7897'
        }

        response = requests.get(
            url,
            headers=headers,
            proxies=proxies,
            timeout=15
        )

        print_info(f"HTTP 状态码: {response.status_code}")

        if response.status_code == 200:
            data = response.json()
            print_success("成功获取模型信息")

            print_info(f"名称: {data.get('name', 'N/A')}")
            print_info(f"显示名称: {data.get('displayName', 'N/A')}")
            print_info(f"版本: {data.get('version', 'N/A')}")
            print_info(f"输入限制: {data.get('inputTokenLimit', 0):,} tokens")
            print_info(f"输出限制: {data.get('outputTokenLimit', 0):,} tokens")

            if "supportedGenerationMethods" in data:
                methods = data["supportedGenerationMethods"]
                print_info(f"支持的方法: {', '.join(methods)}")

            return True, data
        else:
            print_error("无法获取模型信息")
            try:
                error_data = response.json()
                print_error(f"错误: {error_data.get('error', {}).get('message', 'Unknown')}")
            except:
                print_error(f"响应: {response.text[:200]}")
            return False, {}

    except Exception as e:
        print_error(f"发生异常: {str(e)}")
        return False, {"error": str(e)}

def main():
    print_header("🌍 直接测试 Google API（绕过 Sub2API）")

    print_info("代理配置:")
    print_info(f"  HTTP:  http://127.0.0.1:7897")
    print_info(f"  HTTPS: http://127.0.0.1:7897")
    print_info(f"  SOCKS: socks5://127.0.0.1:7897")
    print()
    print_info(f"API Base URL: {GOOGLE_API_BASE}")
    print_info(f"API Key: {GOOGLE_API_KEY[:20]}...{GOOGLE_API_KEY[-10:]}")

    results = []

    for model_id in MODELS:
        print_header(f"测试模型: {model_id}")

        # 测试 1: 获取模型信息
        info_success, info_data = test_model_info(model_id)

        # 测试 2: 生成内容
        gen_success, gen_data = test_google_api_model(model_id)

        results.append({
            "model": model_id,
            "info_available": info_success,
            "generate_available": gen_success,
            "info_data": info_data,
            "generate_data": gen_data
        })

        print("\n" + "-" * 70)

    # 生成对比报告
    print_header("📊 测试结果对比")

    print(f"\n{Colors.BOLD}模型可用性汇总：{Colors.ENDC}\n")
    print(f"{'模型':<35} {'模型信息':<15} {'内容生成':<15} {'结论':<15}")
    print("-" * 70)

    for result in results:
        model = result["model"]
        info = "✓ 可用" if result["info_available"] else "✗ 不可用"
        gen = "✓ 可用" if result["generate_available"] else "✗ 不可用"

        if result["info_available"] and result["generate_available"]:
            conclusion = f"{Colors.OKGREEN}完全可用{Colors.ENDC}"
        elif result["info_available"] or result["generate_available"]:
            conclusion = f"{Colors.WARNING}部分可用{Colors.ENDC}"
        else:
            conclusion = f"{Colors.FAIL}不可用{Colors.ENDC}"

        info_colored = f"{Colors.OKGREEN}{info}{Colors.ENDC}" if result["info_available"] else f"{Colors.FAIL}{info}{Colors.ENDC}"
        gen_colored = f"{Colors.OKGREEN}{gen}{Colors.ENDC}" if result["generate_available"] else f"{Colors.FAIL}{gen}{Colors.ENDC}"

        print(f"{model:<35} {info_colored:<24} {gen_colored:<24} {conclusion}")

    # 保存结果
    output = {
        "timestamp": __import__("datetime").datetime.now().isoformat(),
        "api_base": GOOGLE_API_BASE,
        "proxy": "http://127.0.0.1:7897",
        "results": results
    }

    output_file = "google_api_direct_test.json"
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(output, f, indent=2, ensure_ascii=False)

    print(f"\n详细结果已保存到: {Colors.BOLD}{output_file}{Colors.ENDC}")

    # 最终结论
    print_header("💡 结论")

    available_count = sum(1 for r in results if r["generate_available"])

    if available_count == len(MODELS):
        print_success(f"所有 {len(MODELS)} 个模型在 Google API 中都可用！")
        print()
        print_warning("这意味着问题确实在 Sub2API 的账号配置上：")
        print(f"  • Google API 支持这些模型 ✓")
        print(f"  • Sub2API 需要配置账号来支持这些模型 ✗")
        print()
        print_info("解决方案：")
        print(f"  1. 登录 Sub2API 管理后台")
        print(f"  2. 修改现有 Gemini 账号配置")
        print(f"  3. 将「支持的模型」设置为「所有模型」或添加这些模型")
        print(f"  4. 重新运行测试: python test_gemini_3_series.py")
    elif available_count > 0:
        print_warning(f"{available_count}/{len(MODELS)} 个模型可用")
        print()
        print("可用模型：")
        for r in results:
            if r["generate_available"]:
                print(f"  {Colors.OKGREEN}✓{Colors.ENDC} {r['model']}")
        print()
        print("不可用模型：")
        for r in results:
            if not r["generate_available"]:
                print(f"  {Colors.FAIL}✗{Colors.ENDC} {r['model']}")
                if "error" in r["generate_data"]:
                    print(f"    原因: {r['generate_data']['error'].get('message', 'Unknown')}")
    else:
        print_error("所有模型都不可用！")
        print()
        print_warning("可能的原因：")
        print(f"  • API Key 无效或已过期")
        print(f"  • 代理连接问题")
        print(f"  • Google API 服务问题")
        print(f"  • API Key 没有权限访问这些模型")

if __name__ == "__main__":
    main()
