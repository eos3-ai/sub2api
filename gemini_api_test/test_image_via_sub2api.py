#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试通过 Sub2API 调用 Gemini 3 Pro Image Preview 生成图片
使用 generateContent 端点，直接发送图像描述文本
"""

import requests
import json
import base64
from datetime import datetime
from pathlib import Path

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"
MODEL_ID = "gemini-3-pro-image-preview"

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

def generate_image_via_sub2api(prompt: str, save_image: bool = True):
    """通过 Sub2API 生成图片"""
    print_info(f"图像描述: {prompt}")

    url = f"{API_BASE_URL}/v1beta/models/{MODEL_ID}:generateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    # 使用与 curl 命令相同的格式
    payload = {
        "contents": [{
            "parts": [{
                "text": prompt
            }]
        }]
    }

    print_info(f"请求 URL: {url}")
    print_info(f"请求体: {json.dumps(payload, indent=2, ensure_ascii=False)}")

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
            print_success("请求成功！")

            # 检查响应结构
            print(f"\n{Colors.BOLD}响应结构：{Colors.ENDC}")

            # 显示响应的顶层键
            print_info(f"顶层键: {list(data.keys())}")

            # 检查是否有图片数据
            if "candidates" in data:
                candidates = data["candidates"]
                print_info(f"候选响应数量: {len(candidates)}")

                for idx, candidate in enumerate(candidates):
                    print(f"\n{Colors.BOLD}候选 {idx + 1}:{Colors.ENDC}")

                    content = candidate.get("content", {})
                    parts = content.get("parts", [])

                    print_info(f"  Parts 数量: {len(parts)}")

                    for part_idx, part in enumerate(parts):
                        print(f"  Part {part_idx + 1} 键: {list(part.keys())}")

                        # 检查是否有 inlineData
                        if "inlineData" in part:
                            inline_data = part["inlineData"]
                            mime_type = inline_data.get("mimeType", "unknown")
                            base64_data = inline_data.get("data", "")

                            print_success(f"  ✓ 找到图片数据！")
                            print_info(f"    MIME 类型: {mime_type}")
                            print_info(f"    Base64 数据长度: {len(base64_data)} 字符")

                            if save_image and base64_data:
                                try:
                                    # 解码 base64
                                    image_bytes = base64.b64decode(base64_data)

                                    # 创建输出目录
                                    output_dir = Path("generated_images")
                                    output_dir.mkdir(exist_ok=True)

                                    # 生成文件名
                                    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

                                    # 根据 MIME 类型确定扩展名
                                    if "png" in mime_type.lower():
                                        ext = "png"
                                    elif "jpeg" in mime_type.lower() or "jpg" in mime_type.lower():
                                        ext = "jpg"
                                    elif "webp" in mime_type.lower():
                                        ext = "webp"
                                    else:
                                        ext = "jpg"  # 默认

                                    filename = output_dir / f"gemini_image_{timestamp}_{idx+1}_{part_idx+1}.{ext}"

                                    # 保存图片
                                    with open(filename, "wb") as f:
                                        f.write(image_bytes)

                                    print_success(f"    图片已保存: {filename}")
                                    print_info(f"    文件大小: {len(image_bytes) / 1024:.2f} KB")

                                except Exception as e:
                                    print_error(f"    保存图片失败: {str(e)}")

                        # 检查是否有文本
                        elif "text" in part:
                            text = part["text"]
                            print_info(f"  文本内容: {text[:200]}...")

            # 打印完整响应（前 2000 字符）
            print(f"\n{Colors.BOLD}完整响应（前 2000 字符）：{Colors.ENDC}")
            response_str = json.dumps(data, indent=2, ensure_ascii=False)
            print(response_str[:2000])
            if len(response_str) > 2000:
                print(f"\n{Colors.WARNING}... (响应太长，已截断){Colors.ENDC}")

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

def main():
    print_header("🎨 Gemini 3 Pro Image Preview 图片生成测试（通过 Sub2API）")

    print_info(f"模型: {MODEL_ID}")
    print_info(f"API Base: {API_BASE_URL}")
    print_info("方法: generateContent 端点 + 文本图像描述")
    print()

    # 测试案例（使用中英文）
    test_prompts = [
        "一只可爱的橙色猫咪坐在窗台上看着外面的花园",
        "A serene mountain landscape at sunset with a lake reflecting the sky",
        "未来城市天际线，有飞行汽车和霓虹灯",
    ]

    results = []

    for i, prompt in enumerate(test_prompts, 1):
        print_header(f"测试 {i}/{len(test_prompts)}")

        success, data = generate_image_via_sub2api(prompt, save_image=True)

        results.append({
            "prompt": prompt,
            "success": success,
            "has_image": False,
            "data": data
        })

        # 检查是否真的生成了图片
        if success and data:
            if "candidates" in data:
                for candidate in data["candidates"]:
                    content = candidate.get("content", {})
                    parts = content.get("parts", [])
                    for part in parts:
                        if "inlineData" in part:
                            results[-1]["has_image"] = True
                            break

        print("\n" + "-" * 70 + "\n")

        # 避免请求过快
        if i < len(test_prompts):
            print_info("等待 2 秒后继续下一个测试...")
            import time
            time.sleep(2)

    # 生成测试报告
    print_header("📊 测试结果汇总")

    success_count = sum(1 for r in results if r["success"])
    image_count = sum(1 for r in results if r["has_image"])

    print(f"{Colors.BOLD}测试结果：{Colors.ENDC}")
    print(f"  总测试数: {len(results)}")
    print(f"  {Colors.OKGREEN}请求成功: {success_count}{Colors.ENDC}")
    print(f"  {Colors.OKGREEN}生成图片: {image_count}{Colors.ENDC}")
    print()

    if image_count > 0:
        print_success(f"成功生成了 {image_count} 张图片！")
        print_info("生成的图片保存在 generated_images/ 目录")
        print()
        print_success("✓ gemini-3-pro-image-preview 支持通过 generateContent 端点生成图片")
        print_info("使用方法: 直接发送图像描述文本即可")
    elif success_count > 0:
        print_warning("请求成功，但没有检测到图片数据")
        print_info("可能返回的是文本响应而不是图片")
    else:
        print_error("所有测试都失败了")

    # 保存详细结果
    output = {
        "timestamp": datetime.now().isoformat(),
        "model": MODEL_ID,
        "api_base": API_BASE_URL,
        "method": "generateContent",
        "results": results,
        "summary": {
            "total": len(results),
            "success": success_count,
            "images_generated": image_count
        }
    }

    output_file = "sub2api_image_generation_test.json"
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)

    print(f"\n详细结果已保存到: {Colors.BOLD}{output_file}{Colors.ENDC}")

if __name__ == "__main__":
    main()
