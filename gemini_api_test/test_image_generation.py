#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试 Gemini 3 Pro Image Preview (Nano Banana Pro) 图片生成功能

使用 generateImages 端点测试图片生成
"""

import requests
import json
import base64
from datetime import datetime
from pathlib import Path

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"
MODEL_ID = "gemini-3-pro-image-preview"  # Nano Banana Pro

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

def test_image_generation(prompt: str, save_image: bool = True):
    """测试图片生成"""
    print_header(f"测试图片生成：{MODEL_ID}")

    print_info(f"提示词: {prompt}")

    # 使用 generateImages 端点
    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:generateImages"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "prompt": prompt,
        "number_of_images": 1,
        "aspect_ratio": "1:1",  # 可选：1:1, 16:9, 9:16, 4:3, 3:4
        "safety_filter_level": "block_none",  # 可选：block_none, block_some, block_most
        "person_generation": "allow_adult"  # 可选：allow_adult, allow_all, dont_allow
    }

    print_info(f"请求 URL: {url}")
    print_info(f"请求参数: {json.dumps(payload, indent=2, ensure_ascii=False)}")

    try:
        print_info("发送图片生成请求...")
        start_time = datetime.now()

        response = requests.post(
            url,
            headers=headers,
            json=payload,
            timeout=120  # 图片生成可能需要较长时间
        )

        elapsed = (datetime.now() - start_time).total_seconds()

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f} 秒")

        if response.status_code == 200:
            data = response.json()
            print_success("图片生成成功！")

            # 检查响应结构
            if "generatedImages" in data:
                images = data["generatedImages"]
                print_info(f"生成了 {len(images)} 张图片")

                for idx, img in enumerate(images):
                    print(f"\n{Colors.BOLD}图片 {idx + 1}:{Colors.ENDC}")

                    # 检查图片数据
                    if "image" in img:
                        image_data = img["image"]

                        # 可能是 base64 编码
                        if "bytesBase64Encoded" in image_data:
                            base64_data = image_data["bytesBase64Encoded"]
                            print_info(f"Base64 数据长度: {len(base64_data)} 字符")

                            if save_image:
                                # 保存图片
                                image_bytes = base64.b64decode(base64_data)

                                # 创建输出目录
                                output_dir = Path("generated_images")
                                output_dir.mkdir(exist_ok=True)

                                # 生成文件名
                                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                                filename = output_dir / f"gemini_image_{timestamp}_{idx+1}.png"

                                # 保存
                                with open(filename, "wb") as f:
                                    f.write(image_bytes)

                                print_success(f"图片已保存: {filename}")
                                print_info(f"图片大小: {len(image_bytes) / 1024:.2f} KB")

                        # 可能是 URL
                        elif "uri" in image_data:
                            uri = image_data["uri"]
                            print_info(f"图片 URL: {uri}")

                        # 检查其他字段
                        if "mimeType" in image_data:
                            print_info(f"MIME 类型: {image_data['mimeType']}")

                    # 检查生成信息
                    if "generationSeed" in img:
                        print_info(f"生成种子: {img['generationSeed']}")

                    if "raiFilteredReason" in img:
                        print_warning(f"安全过滤: {img['raiFilteredReason']}")

            # 打印完整响应（调试用）
            print(f"\n{Colors.BOLD}完整响应：{Colors.ENDC}")
            print(json.dumps(data, indent=2, ensure_ascii=False)[:1000])

            return True, data

        else:
            print_error(f"图片生成失败 (HTTP {response.status_code})")

            try:
                error_data = response.json()
                error_msg = error_data.get("error", {}).get("message", "Unknown error")
                print_error(f"错误信息: {error_msg}")

                print(f"\n{Colors.BOLD}错误响应：{Colors.ENDC}")
                print(json.dumps(error_data, indent=2, ensure_ascii=False))
            except:
                print_error(f"响应内容: {response.text[:500]}")

            return False, None

    except requests.exceptions.Timeout:
        print_error("请求超时（120秒）")
        return False, None
    except Exception as e:
        print_error(f"发生异常: {str(e)}")
        import traceback
        traceback.print_exc()
        return False, None

def test_text_generation():
    """测试文本生成（作为对照）"""
    print_header("测试文本生成功能（对照测试）")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:generateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{"text": "Describe a sunset in one sentence."}]
        }],
        "generationConfig": {
            "maxOutputTokens": 100
        }
    }

    print_info(f"请求 URL: {url}")

    try:
        response = requests.post(url, headers=headers, json=payload, timeout=30)

        print_info(f"HTTP 状态码: {response.status_code}")

        if response.status_code == 200:
            print_success("文本生成成功！")
            data = response.json()

            if "candidates" in data and len(data["candidates"]) > 0:
                content = data["candidates"][0].get("content", {})
                parts = content.get("parts", [])
                if len(parts) > 0:
                    text = parts[0].get("text", "")
                    print_info(f"生成文本: {text}")

            return True
        else:
            print_error("文本生成失败")
            print_error(f"响应: {response.text[:300]}")
            return False

    except Exception as e:
        print_error(f"异常: {str(e)}")
        return False

def main():
    print_header("🎨 Gemini 3 Pro Image Preview 图片生成测试")

    print_info(f"模型: {MODEL_ID} (Nano Banana Pro)")
    print_info(f"API Base: {API_BASE_URL}")
    print()

    # 测试案例
    test_prompts = [
        "A serene mountain landscape at sunset with a lake reflecting the sky",
        "一只可爱的橙色猫咪坐在窗台上看着外面的花园",
        "Futuristic city skyline with flying cars and neon lights"
    ]

    results = []

    # 首先测试文本生成功能
    print_header("🔍 前置检查：测试文本生成")
    text_success = test_text_generation()

    if not text_success:
        print_error("\n文本生成失败，可能账号配置有问题")
        print_info("建议先运行: python diagnose_models.py")
        return

    print()

    # 测试图片生成
    for i, prompt in enumerate(test_prompts, 1):
        print_header(f"测试 {i}/{len(test_prompts)}")

        success, data = test_image_generation(prompt, save_image=True)

        results.append({
            "prompt": prompt,
            "success": success,
            "data": data
        })

        print("\n" + "-" * 70 + "\n")

        # 避免请求过快
        if i < len(test_prompts):
            print_info("等待 3 秒后继续下一个测试...")
            import time
            time.sleep(3)

    # 生成测试报告
    print_header("📊 测试结果汇总")

    success_count = sum(1 for r in results if r["success"])

    print(f"{Colors.BOLD}图片生成测试结果：{Colors.ENDC}")
    print(f"  总测试数: {len(results)}")
    print(f"  {Colors.OKGREEN}成功: {success_count}{Colors.ENDC}")
    print(f"  {Colors.FAIL}失败: {len(results) - success_count}{Colors.ENDC}")
    print()

    if success_count > 0:
        print_success("模型支持图片生成功能！")
        print_info("生成的图片保存在 generated_images/ 目录")
    else:
        print_error("图片生成功能测试失败")
        print()
        print_info("可能的原因：")
        print("  1. 模型不支持 generateImages 端点（可能仅支持文本生成）")
        print("  2. API 端点或参数格式错误")
        print("  3. 需要特殊的认证或配置")
        print()
        print_info("Nano Banana Pro 可能主要是文本模型，图片生成功能需要确认")

    # 保存详细结果
    output = {
        "timestamp": datetime.now().isoformat(),
        "model": MODEL_ID,
        "text_generation_test": text_success,
        "image_generation_tests": results,
        "summary": {
            "total": len(results),
            "success": success_count,
            "failed": len(results) - success_count
        }
    }

    output_file = "image_generation_test_results.json"
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)

    print(f"\n详细结果已保存到: {Colors.BOLD}{output_file}{Colors.ENDC}")

if __name__ == "__main__":
    main()
