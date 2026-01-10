#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Gemini 模型可用性诊断工具

帮助诊断为什么某些模型返回 503 错误
"""

import requests
import json
from typing import Dict, List, Tuple

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"

# Gemini 3 系列模型
GEMINI_3_MODELS = [
    "gemini-3-pro-preview",
    "gemini-3-flash-preview",
    "gemini-3-pro-image-preview"
]

class Colors:
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    OKCYAN = '\033[96m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text: str):
    print(f"\n{Colors.BOLD}{'='*70}")
    print(f"{text}")
    print(f"{'='*70}{Colors.ENDC}\n")

def check_model_availability(model_id: str) -> Tuple[bool, Dict]:
    """检查单个模型的可用性"""
    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    payload = {
        "contents": [{
            "role": "user",
            "parts": [{"text": "test"}]
        }],
        "generationConfig": {
            "maxOutputTokens": 10
        }
    }

    try:
        response = requests.post(url, headers=headers, json=payload, timeout=10)

        result = {
            "status_code": response.status_code,
            "available": response.status_code == 200,
            "error": None
        }

        if response.status_code != 200:
            try:
                error_data = response.json()
                result["error"] = error_data.get("error", {}).get("message", "Unknown error")
            except:
                result["error"] = response.text[:200]

        return result["available"], result

    except Exception as e:
        return False, {
            "status_code": 0,
            "available": False,
            "error": str(e)
        }

def get_model_details(model_id: str) -> Dict:
    """获取模型详细信息"""
    url = f"{API_BASE_URL}/v1beta/models/{model_id}"
    headers = {
        "Authorization": f"Bearer {API_KEY}"
    }

    try:
        response = requests.get(url, headers=headers, timeout=10)
        if response.status_code == 200:
            return response.json()
        return None
    except:
        return None

def diagnose():
    """诊断所有 Gemini 3 模型"""
    print_header("🔍 Gemini 3.x 系列模型诊断工具")

    print(f"{Colors.OKCYAN}正在检查 {len(GEMINI_3_MODELS)} 个 Gemini 3 系列模型...{Colors.ENDC}\n")

    results = []

    for model_id in GEMINI_3_MODELS:
        print(f"检查模型: {Colors.BOLD}{model_id}{Colors.ENDC}")

        # 检查模型信息
        model_info = get_model_details(model_id)
        if model_info:
            print(f"  ✓ 模型存在于列表中")
            display_name = model_info.get("displayName", "N/A")
            print(f"    名称: {display_name}")
        else:
            print(f"  {Colors.WARNING}⚠ 无法获取模型信息{Colors.ENDC}")

        # 检查可用性
        available, result = check_model_availability(model_id)

        if available:
            print(f"  {Colors.OKGREEN}✓ 模型可用 (HTTP {result['status_code']}){Colors.ENDC}")
        else:
            print(f"  {Colors.FAIL}✗ 模型不可用 (HTTP {result['status_code']}){Colors.ENDC}")
            if result['error']:
                print(f"    错误: {Colors.FAIL}{result['error']}{Colors.ENDC}")

                # 分析错误原因
                if "No available Gemini accounts" in result['error']:
                    print(f"    {Colors.WARNING}原因: 转发服务未配置该模型的可用账号{Colors.ENDC}")
                elif "quota" in result['error'].lower():
                    print(f"    {Colors.WARNING}原因: API 配额不足{Colors.ENDC}")
                elif "permission" in result['error'].lower():
                    print(f"    {Colors.WARNING}原因: 权限不足{Colors.ENDC}")

        results.append({
            "model": model_id,
            "available": available,
            "status_code": result.get("status_code"),
            "error": result.get("error")
        })

        print()

    # 生成诊断报告
    print_header("📊 诊断报告")

    available_count = sum(1 for r in results if r["available"])
    unavailable_count = len(results) - available_count

    print(f"总模型数: {len(results)}")
    print(f"{Colors.OKGREEN}可用: {available_count}{Colors.ENDC}")
    print(f"{Colors.FAIL}不可用: {unavailable_count}{Colors.ENDC}\n")

    if unavailable_count > 0:
        print(f"{Colors.BOLD}不可用模型详情：{Colors.ENDC}\n")

        for r in results:
            if not r["available"]:
                print(f"• {Colors.FAIL}{r['model']}{Colors.ENDC}")
                print(f"  状态码: {r['status_code']}")
                print(f"  错误: {r['error']}\n")

        print(f"{Colors.BOLD}推荐操作：{Colors.ENDC}\n")

        # 检查错误类型
        has_account_error = any("No available Gemini accounts" in (r.get("error") or "") for r in results if not r["available"])

        if has_account_error:
            print(f"{Colors.OKCYAN}1. 问题原因{Colors.ENDC}")
            print(f"   转发服务识别了这些模型，但没有配置可用的 Gemini 账号。\n")

            print(f"{Colors.OKCYAN}2. 解决方案{Colors.ENDC}")
            print(f"   在 sub2api 后台添加支持这些模型的 Gemini 账号：\n")

            print(f"   {Colors.BOLD}方法 A: 使用 Web UI（推荐）{Colors.ENDC}")
            print(f"   • 访问管理后台：http://localhost:8080")
            print(f"   • 进入「账号管理」页面")
            print(f"   • 点击「添加账号」")
            print(f"   • 选择平台：Gemini")
            print(f"   • 账号类型：API Key")
            print(f"   • 填写 API Key：AIzaSyD76zD8PA1RsEBeCaN-PhqpWj-rDfpXDyk")
            print(f"   • 支持的模型：选择所有 Gemini 3 系列模型（或选择「支持所有模型」）")
            print(f"   • 保存并启用\n")

            print(f"   {Colors.BOLD}方法 B: 检查现有账号配置{Colors.ENDC}")
            print(f"   • 如果已有 Gemini 账号，检查其「支持的模型」列表")
            print(f"   • 确保包含：")
            for model in GEMINI_3_MODELS:
                available = any(r["model"] == model and r["available"] for r in results)
                status = f"{Colors.OKGREEN}✓{Colors.ENDC}" if available else f"{Colors.FAIL}✗{Colors.ENDC}"
                print(f"     {status} {model}")
            print()

            print(f"{Colors.OKCYAN}3. 验证配置{Colors.ENDC}")
            print(f"   添加账号后，重新运行此诊断工具：")
            print(f"   {Colors.BOLD}python diagnose_models.py{Colors.ENDC}\n")

        print(f"{Colors.OKCYAN}4. 更多帮助{Colors.ENDC}")
        print(f"   查看详细配置指南：")
        print(f"   {Colors.BOLD}cat check_account_config.md{Colors.ENDC}\n")

    else:
        print(f"{Colors.OKGREEN}✓ 所有模型都可用！{Colors.ENDC}\n")

    # 保存诊断结果
    output = {
        "timestamp": __import__("datetime").datetime.now().isoformat(),
        "summary": {
            "total": len(results),
            "available": available_count,
            "unavailable": unavailable_count
        },
        "results": results
    }

    output_file = "diagnosis_result.json"
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(output, f, indent=2, ensure_ascii=False)

    print(f"诊断结果已保存到: {Colors.BOLD}{output_file}{Colors.ENDC}\n")

if __name__ == "__main__":
    diagnose()
