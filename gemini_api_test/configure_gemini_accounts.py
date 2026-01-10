#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Sub2API Gemini 账号自动配置工具

功能：
1. 列出所有 Gemini 账号
2. 更新账号配置以支持所有 Gemini 3.x 模型
3. 验证配置是否生效
"""

import requests
import json
import sys
from typing import List, Dict, Optional

# 配置
API_BASE = "https://router.aitokencloud.com"
ADMIN_API_KEY = "admin-7d334c231566cd7a4e8bca75e04d5833c1af3c0f0b2350562d0add761ee6812e"

# 目标模型
TARGET_MODELS = [
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

def list_gemini_accounts() -> List[Dict]:
    """列出所有 Gemini 账号"""
    print_info("正在查询 Gemini 账号...")

    url = f"{API_BASE}/api/v1/admin/accounts"
    headers = {"x-api-key": ADMIN_API_KEY}
    params = {"platform": "gemini"}

    try:
        response = requests.get(url, headers=headers, params=params, timeout=10)

        if response.status_code != 200:
            print_error(f"查询失败 (HTTP {response.status_code})")
            print_error(f"响应: {response.text[:300]}")
            return []

        response_data = response.json()
        # API 返回结构: {"code": 0, "data": {"items": [...], "total": 1}}
        data = response_data.get("data", {})
        accounts = data.get("items", [])

        print_success(f"找到 {len(accounts)} 个 Gemini 账号\n")

        for acc in accounts:
            print(f"{Colors.BOLD}账号 ID: {acc['id']}{Colors.ENDC}")
            print(f"  名称: {acc.get('name', 'N/A')}")
            print(f"  状态: {acc.get('status', 'N/A')}")
            print(f"  类型: {acc.get('type', 'N/A')}")

            # 显示当前的 model_mapping
            credentials = acc.get('credentials', {})
            model_mapping = credentials.get('model_mapping')

            if model_mapping is None:
                print(f"  {Colors.OKGREEN}模型映射: 支持所有模型 (无限制){Colors.ENDC}")
            elif isinstance(model_mapping, dict) and len(model_mapping) == 0:
                print(f"  {Colors.OKGREEN}模型映射: 支持所有模型 (空映射){Colors.ENDC}")
            elif isinstance(model_mapping, dict):
                print(f"  {Colors.WARNING}模型映射: 仅支持 {len(model_mapping)} 个模型{Colors.ENDC}")
                for model_name in model_mapping.keys():
                    status = f"{Colors.OKGREEN}✓{Colors.ENDC}" if model_name in TARGET_MODELS else "•"
                    print(f"    {status} {model_name}")
            else:
                print(f"  {Colors.WARNING}模型映射: {model_mapping}{Colors.ENDC}")

            print()

        return accounts

    except requests.exceptions.RequestException as e:
        print_error(f"网络请求失败: {str(e)}")
        return []
    except Exception as e:
        print_error(f"发生异常: {str(e)}")
        return []

def get_account_models(account_id: int) -> List[str]:
    """获取账号支持的模型列表"""
    url = f"{API_BASE}/api/v1/admin/accounts/{account_id}/models"
    headers = {"x-api-key": ADMIN_API_KEY}

    try:
        response = requests.get(url, headers=headers, timeout=10)

        if response.status_code != 200:
            print_warning(f"无法获取账号 {account_id} 的模型列表")
            return []

        models = response.json()
        return [m.get('id', '') for m in models if isinstance(m, dict)]

    except Exception as e:
        print_warning(f"获取模型列表失败: {str(e)}")
        return []

def update_account_to_support_all_models(account_id: int, account_name: str, current_credentials: dict) -> bool:
    """更新账号配置以支持所有模型"""
    print(f"\n{Colors.BOLD}正在配置账号 {account_id} ({account_name})...{Colors.ENDC}")

    url = f"{API_BASE}/api/v1/admin/accounts/{account_id}"
    headers = {
        "x-api-key": ADMIN_API_KEY,
        "Content-Type": "application/json"
    }

    # 方案：保留现有 credentials，只修改 model_mapping
    new_credentials = current_credentials.copy()
    new_credentials["model_mapping"] = None

    payload = {
        "credentials": new_credentials
    }

    try:
        print_info("  发送更新请求...")
        response = requests.put(url, headers=headers, json=payload, timeout=10)

        if response.status_code == 200:
            print_success(f"  账号 {account_id} 配置已更新为支持所有模型")
            return True
        else:
            print_error(f"  更新失败 (HTTP {response.status_code})")
            print_error(f"  响应: {response.text[:300]}")
            return False

    except Exception as e:
        print_error(f"  更新失败: {str(e)}")
        return False

def verify_account_configuration(account_id: int, account_name: str) -> bool:
    """验证账号配置是否生效"""
    print(f"\n{Colors.BOLD}验证账号 {account_id} ({account_name}) 的配置...{Colors.ENDC}")

    # 获取支持的模型列表
    models = get_account_models(account_id)

    if not models:
        print_warning("  无法获取模型列表，跳过验证")
        return False

    print_info(f"  账号支持 {len(models)} 个模型")

    # 检查目标模型是否都支持
    found = [m for m in TARGET_MODELS if m in models]
    missing = [m for m in TARGET_MODELS if m not in models]

    print(f"\n  {Colors.BOLD}Gemini 3.x 模型支持情况：{Colors.ENDC}")
    for model in TARGET_MODELS:
        if model in found:
            print(f"    {Colors.OKGREEN}✓ {model}{Colors.ENDC}")
        else:
            print(f"    {Colors.FAIL}✗ {model}{Colors.ENDC}")

    if len(found) == len(TARGET_MODELS):
        print_success(f"\n  所有 {len(TARGET_MODELS)} 个目标模型都已支持！")
        return True
    else:
        print_warning(f"\n  仅支持 {len(found)}/{len(TARGET_MODELS)} 个目标模型")
        if missing:
            print_warning(f"  缺少: {', '.join(missing)}")
        return False

def main():
    """主函数"""
    print_header("🔧 Sub2API Gemini 账号自动配置工具")

    if not ADMIN_API_KEY:
        print_error("错误: 未设置 ADMIN_API_KEY")
        sys.exit(1)

    print_info(f"API Base URL: {API_BASE}")
    print_info(f"Admin API Key: {ADMIN_API_KEY[:30]}...{ADMIN_API_KEY[-10:]}")
    print()

    # 步骤 1: 列出所有 Gemini 账号
    print_header("📋 步骤 1: 查询现有 Gemini 账号")
    accounts = list_gemini_accounts()

    if not accounts:
        print_error("未找到任何 Gemini 账号")
        print()
        print_info("请先在 Sub2API 后台添加 Gemini 账号：")
        print_info("  1. 访问: http://localhost:8080/admin/accounts")
        print_info("  2. 点击「添加账号」")
        print_info("  3. 选择平台: Gemini")
        print_info("  4. 填写 API Key 并保存")
        sys.exit(0)

    # 步骤 2: 更新每个账号的配置
    print_header("⚙️  步骤 2: 更新账号配置")

    updated_accounts = []
    failed_accounts = []

    for acc in accounts:
        account_id = acc['id']
        account_name = acc.get('name', f'Account {account_id}')

        # 检查当前配置
        credentials = acc.get('credentials', {})
        model_mapping = credentials.get('model_mapping')

        # 如果已经是支持所有模型，跳过
        if model_mapping is None or (isinstance(model_mapping, dict) and len(model_mapping) == 0):
            print_info(f"账号 {account_id} ({account_name}) 已配置为支持所有模型，跳过")
            updated_accounts.append(account_id)
            continue

        # 更新配置（传入完整的 credentials）
        if update_account_to_support_all_models(account_id, account_name, credentials):
            updated_accounts.append(account_id)
        else:
            failed_accounts.append(account_id)

    # 步骤 3: 验证配置
    print_header("✅ 步骤 3: 验证配置")

    verification_results = []

    for account_id in updated_accounts:
        account_name = next((acc.get('name', f'Account {account_id}')
                           for acc in accounts if acc['id'] == account_id),
                          f'Account {account_id}')

        success = verify_account_configuration(account_id, account_name)
        verification_results.append({
            'account_id': account_id,
            'account_name': account_name,
            'verified': success
        })

    # 生成最终报告
    print_header("📊 配置结果汇总")

    print(f"{Colors.BOLD}账号处理结果：{Colors.ENDC}")
    print(f"  总计: {len(accounts)} 个账号")
    print(f"  {Colors.OKGREEN}已更新: {len(updated_accounts)} 个{Colors.ENDC}")
    if failed_accounts:
        print(f"  {Colors.FAIL}失败: {len(failed_accounts)} 个{Colors.ENDC}")
    print()

    print(f"{Colors.BOLD}配置验证结果：{Colors.ENDC}")
    verified_count = sum(1 for r in verification_results if r['verified'])
    print(f"  {Colors.OKGREEN}验证通过: {verified_count}/{len(verification_results)} 个账号{Colors.ENDC}")

    if verified_count < len(verification_results):
        print()
        print(f"{Colors.WARNING}部分账号验证失败，可能需要：{Colors.ENDC}")
        print(f"  • 等待几秒钟让配置生效")
        print(f"  • 检查账号状态是否为 active")
        print(f"  • 确认 API Key 有效")

    # 下一步指引
    print()
    print_header("🎯 下一步")

    if verified_count > 0:
        print_success("配置已完成！现在可以测试所有 Gemini 3.x 模型了。\n")

        print(f"{Colors.BOLD}运行诊断工具验证：{Colors.ENDC}")
        print(f"  cd gemini_api_test")
        print(f"  python diagnose_models.py\n")

        print(f"{Colors.BOLD}运行完整测试套件：{Colors.ENDC}")
        print(f"  python test_gemini_3_series.py\n")

        print(f"{Colors.BOLD}预期结果：{Colors.ENDC}")
        print(f"  • 所有 3 个模型显示为可用")
        print(f"  • 所有 12 个测试通过 (4 测试 × 3 模型)")
    else:
        print_error("配置未能完成，请检查上述错误信息。\n")

        print(f"{Colors.BOLD}故障排查步骤：{Colors.ENDC}")
        print(f"  1. 检查 Admin API Key 是否有效")
        print(f"  2. 确认账号状态为 active")
        print(f"  3. 查看 sub2api 后端日志")
        print(f"  4. 手动访问管理后台验证配置")

    print()

if __name__ == "__main__":
    main()
