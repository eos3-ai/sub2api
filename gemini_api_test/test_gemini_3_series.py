#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Gemini 3.x 系列模型对比测试

测试三个 Gemini 3.x 模型：
1. gemini-3-pro-preview - 支持推理模式
2. gemini-3-flash-preview - 快速版本
3. gemini-3-pro-image-preview - 图像生成
"""

import requests
import json
import time
from datetime import datetime
from typing import Dict, Any, Tuple, List

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"
TIMEOUT = 60  # 增加超时时间

# 测试模型列表
MODELS = [
    {
        "id": "gemini-3-pro-preview",
        "name": "Gemini 3 Pro Preview",
        "description": "支持推理模式的旗舰模型",
        "supports_thinking": True
    },
    {
        "id": "gemini-3-flash-preview",
        "name": "Gemini 3 Flash Preview",
        "description": "快速响应的轻量级模型",
        "supports_thinking": False
    },
    {
        "id": "gemini-3-pro-image-preview",
        "name": "Nano Banana Pro",
        "description": "图像生成专用模型",
        "supports_thinking": False
    }
]

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
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*70}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*70}{Colors.ENDC}\n")

def print_success(text: str):
    """打印成功信息"""
    print(f"{Colors.OKGREEN}✓ {text}{Colors.ENDC}")

def print_error(text: str):
    """打印错误信息"""
    print(f"{Colors.FAIL}✗ {text}{Colors.ENDC}")

def print_info(text: str):
    """打印信息"""
    print(f"{Colors.OKCYAN}ℹ {text}{Colors.ENDC}")

def print_warning(text: str):
    """打印警告"""
    print(f"{Colors.WARNING}⚠ {text}{Colors.ENDC}")

def get_headers() -> Dict[str, str]:
    """获取请求头"""
    return {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

def test_model_basic(model_id: str, model_name: str) -> Tuple[bool, Dict[str, Any]]:
    """
    测试基本文本生成功能
    """
    print_info(f"测试模型: {model_name} ({model_id})")

    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "请用一句话解释什么是量子计算。"}]
        }],
        "generationConfig": {
            "maxOutputTokens": 200,
            "temperature": 0.7
        }
    }

    start_time = time.time()

    try:
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

            # 提取生成的文本
            generated_text = ""
            if "candidates" in data and len(data["candidates"]) > 0:
                candidate = data["candidates"][0]
                if "content" in candidate and "parts" in candidate["content"]:
                    parts = candidate["content"]["parts"]
                    if len(parts) > 0:
                        generated_text = parts[0].get("text", "")

            # 提取用量信息
            usage = data.get("usageMetadata", {})
            prompt_tokens = usage.get("promptTokenCount", 0)
            output_tokens = usage.get("candidatesTokenCount", 0)
            total_tokens = usage.get("totalTokenCount", 0)
            thinking_tokens = usage.get("thoughtsTokenCount", 0)

            print_success("成功生成内容")
            if generated_text:
                print_info(f"生成内容: {generated_text[:100]}...")
            print_info(f"Token 使用: 输入={prompt_tokens}, 输出={output_tokens}, 总计={total_tokens}")
            if thinking_tokens > 0:
                print_warning(f"推理 Token: {thinking_tokens} (该模型使用了推理模式)")

            return True, {
                "status": "success",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "generated_text": generated_text,
                "prompt_tokens": prompt_tokens,
                "output_tokens": output_tokens,
                "total_tokens": total_tokens,
                "thinking_tokens": thinking_tokens
            }
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            print_error(f"错误信息: {response.text[:300]}")
            return False, {
                "status": "failed",
                "status_code": response.status_code,
                "elapsed": elapsed,
                "error": response.text[:300]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_model_reasoning(model_id: str, model_name: str) -> Tuple[bool, Dict[str, Any]]:
    """
    测试推理能力（复杂问题）
    """
    print_info(f"测试推理能力: {model_name}")

    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "一个班级有30个学生，其中2/3是女生。如果再加入6个男生，那么男生占总人数的百分比是多少？请详细解释计算过程。"}]
        }],
        "generationConfig": {
            "maxOutputTokens": 500,
            "temperature": 0.7
        }
    }

    start_time = time.time()

    try:
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

            # 提取生成的文本
            generated_text = ""
            if "candidates" in data and len(data["candidates"]) > 0:
                candidate = data["candidates"][0]
                if "content" in candidate and "parts" in candidate["content"]:
                    parts = candidate["content"]["parts"]
                    if len(parts) > 0:
                        generated_text = parts[0].get("text", "")

            # 提取用量信息
            usage = data.get("usageMetadata", {})
            thinking_tokens = usage.get("thoughtsTokenCount", 0)

            print_success("成功生成推理答案")
            if generated_text:
                print_info(f"推理结果: {generated_text[:150]}...")

            if thinking_tokens > 0:
                print_warning(f"使用推理 Token: {thinking_tokens}")

            return True, {
                "status": "success",
                "elapsed": elapsed,
                "generated_text": generated_text,
                "thinking_tokens": thinking_tokens,
                "has_reasoning": thinking_tokens > 0
            }
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False, {
                "status": "failed",
                "elapsed": elapsed,
                "error": response.text[:300]
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_model_creative(model_id: str, model_name: str) -> Tuple[bool, Dict[str, Any]]:
    """
    测试创意生成能力
    """
    print_info(f"测试创意生成: {model_name}")

    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "写一首关于春天的四行诗，要求有韵律感。"}]
        }],
        "generationConfig": {
            "maxOutputTokens": 300,
            "temperature": 0.9
        }
    }

    start_time = time.time()

    try:
        response = requests.post(
            url,
            headers=get_headers(),
            json=payload,
            timeout=TIMEOUT
        )
        elapsed = time.time() - start_time

        if response.status_code == 200:
            data = response.json()

            # 提取生成的文本
            generated_text = ""
            if "candidates" in data and len(data["candidates"]) > 0:
                candidate = data["candidates"][0]
                if "content" in candidate and "parts" in candidate["content"]:
                    parts = candidate["content"]["parts"]
                    if len(parts) > 0:
                        generated_text = parts[0].get("text", "")

            print_success("成功生成创意内容")
            if generated_text:
                print_info(f"创意作品:\n{generated_text}")

            return True, {
                "status": "success",
                "elapsed": elapsed,
                "generated_text": generated_text
            }
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False, {
                "status": "failed",
                "elapsed": elapsed
            }

    except Exception as e:
        elapsed = time.time() - start_time
        print_error(f"发生异常: {str(e)}")
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def test_model_speed(model_id: str, model_name: str) -> Tuple[bool, Dict[str, Any]]:
    """
    测试响应速度（简单问题）
    """
    print_info(f"测试响应速度: {model_name}")

    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"

    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "Hello, how are you?"}]
        }],
        "generationConfig": {
            "maxOutputTokens": 50
        }
    }

    start_time = time.time()

    try:
        response = requests.post(
            url,
            headers=get_headers(),
            json=payload,
            timeout=TIMEOUT
        )
        elapsed = time.time() - start_time

        if response.status_code == 200:
            print_success(f"响应时间: {elapsed:.3f}秒")
            return True, {
                "status": "success",
                "elapsed": elapsed
            }
        else:
            return False, {
                "status": "failed",
                "elapsed": elapsed
            }

    except Exception as e:
        elapsed = time.time() - start_time
        return False, {
            "status": "error",
            "elapsed": elapsed,
            "error": str(e)
        }

def compare_models():
    """对比所有模型"""
    print(f"\n{Colors.BOLD}{'='*70}")
    print(f"Gemini 3.x 系列模型对比测试")
    print(f"{'='*70}{Colors.ENDC}\n")

    print_info(f"API Base URL: {API_BASE_URL}")
    print_info(f"测试时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print_info(f"测试模型数量: {len(MODELS)}")

    all_results = {
        "test_time": datetime.now().isoformat(),
        "api_base_url": API_BASE_URL,
        "models": []
    }

    for model_info in MODELS:
        model_id = model_info["id"]
        model_name = model_info["name"]

        print_header(f"测试模型: {model_name}")
        print_info(f"模型 ID: {model_id}")
        print_info(f"描述: {model_info['description']}")

        model_results = {
            "model_id": model_id,
            "model_name": model_name,
            "description": model_info["description"],
            "tests": {}
        }

        # 测试 1: 基本文本生成
        print(f"\n{Colors.OKCYAN}【测试 1/4】基本文本生成{Colors.ENDC}")
        success, result = test_model_basic(model_id, model_name)
        model_results["tests"]["basic"] = {"success": success, "result": result}
        time.sleep(1)

        # 测试 2: 推理能力
        print(f"\n{Colors.OKCYAN}【测试 2/4】推理能力测试{Colors.ENDC}")
        success, result = test_model_reasoning(model_id, model_name)
        model_results["tests"]["reasoning"] = {"success": success, "result": result}
        time.sleep(1)

        # 测试 3: 创意生成
        print(f"\n{Colors.OKCYAN}【测试 3/4】创意生成测试{Colors.ENDC}")
        success, result = test_model_creative(model_id, model_name)
        model_results["tests"]["creative"] = {"success": success, "result": result}
        time.sleep(1)

        # 测试 4: 响应速度
        print(f"\n{Colors.OKCYAN}【测试 4/4】响应速度测试{Colors.ENDC}")
        success, result = test_model_speed(model_id, model_name)
        model_results["tests"]["speed"] = {"success": success, "result": result}

        all_results["models"].append(model_results)

        print("\n" + "-" * 70)
        time.sleep(2)

    # 生成对比报告
    print_header("测试结果对比")

    # 对比表格
    print(f"\n{Colors.BOLD}响应时间对比（秒）：{Colors.ENDC}")
    print(f"{'模型':<30} {'基本生成':<12} {'推理测试':<12} {'创意生成':<12} {'速度测试':<12}")
    print("-" * 70)

    for model_result in all_results["models"]:
        model_name = model_result["model_name"][:28]
        basic_time = model_result["tests"]["basic"]["result"].get("elapsed", 0)
        reasoning_time = model_result["tests"]["reasoning"]["result"].get("elapsed", 0)
        creative_time = model_result["tests"]["creative"]["result"].get("elapsed", 0)
        speed_time = model_result["tests"]["speed"]["result"].get("elapsed", 0)

        print(f"{model_name:<30} {basic_time:<12.2f} {reasoning_time:<12.2f} {creative_time:<12.2f} {speed_time:<12.2f}")

    # Token 使用对比
    print(f"\n{Colors.BOLD}Token 使用对比（基本生成测试）：{Colors.ENDC}")
    print(f"{'模型':<30} {'输入':<10} {'输出':<10} {'总计':<10} {'推理':<10}")
    print("-" * 70)

    for model_result in all_results["models"]:
        model_name = model_result["model_name"][:28]
        basic_result = model_result["tests"]["basic"]["result"]
        prompt = basic_result.get("prompt_tokens", 0)
        output = basic_result.get("output_tokens", 0)
        total = basic_result.get("total_tokens", 0)
        thinking = basic_result.get("thinking_tokens", 0)

        print(f"{model_name:<30} {prompt:<10} {output:<10} {total:<10} {thinking:<10}")

    # 推理能力对比
    print(f"\n{Colors.BOLD}推理能力对比：{Colors.ENDC}")
    for model_result in all_results["models"]:
        model_name = model_result["model_name"]
        reasoning_result = model_result["tests"]["reasoning"]["result"]
        thinking_tokens = reasoning_result.get("thinking_tokens", 0)
        has_reasoning = reasoning_result.get("has_reasoning", False)

        if has_reasoning:
            print_success(f"{model_name}: 支持推理模式（使用 {thinking_tokens} 个推理 Token）")
        else:
            print_info(f"{model_name}: 标准模式（无推理 Token）")

    # 保存结果
    output_file = "gemini_3_series_comparison.json"
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(all_results, f, indent=2, ensure_ascii=False)
        print(f"\n{Colors.OKGREEN}✓ 详细测试结果已保存到: {output_file}{Colors.ENDC}")
    except Exception as e:
        print_error(f"保存结果失败: {str(e)}")

    # 最终总结
    print(f"\n{Colors.BOLD}{'='*70}")
    print(f"测试总结")
    print(f"{'='*70}{Colors.ENDC}\n")

    for model_result in all_results["models"]:
        model_name = model_result["model_name"]
        tests = model_result["tests"]

        passed = sum(1 for t in tests.values() if t["success"])
        total = len(tests)

        if passed == total:
            print_success(f"{model_name}: 所有测试通过 ({passed}/{total})")
        else:
            print_warning(f"{model_name}: {passed}/{total} 测试通过")

    print()

def main():
    """主函数"""
    try:
        compare_models()
        return 0
    except KeyboardInterrupt:
        print(f"\n\n{Colors.WARNING}测试被用户中断{Colors.ENDC}")
        return 1
    except Exception as e:
        print_error(f"发生未预期的错误: {str(e)}")
        return 1

if __name__ == "__main__":
    exit(main())
