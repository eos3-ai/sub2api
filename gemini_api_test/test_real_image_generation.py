#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试 Imagen 4 图片生成功能
"""

import requests
import json
import base64
from datetime import datetime
from pathlib import Path

# 配置
API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"

# 测试的模型
MODELS = [
    {
        "id": "imagen-4.0-fast-generate-001",
        "name": "Imagen 4 Fast",
        "description": "快速生成"
    },
    {
        "id": "gemini-2.0-flash-exp-image-generation",
        "name": "Gemini 2.0 Flash Image Generation",
        "description": "实验性图片生成"
    }
]

# 颜色输出
class Colors:
    HEADER = '\033[95m'
    OKGREEN = '\033[92m'
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

def test_imagen_4(model_id: str, prompt: str):
    """测试 Imagen 4 图片生成"""
    print_info(f"模型: {model_id}")
    print_info(f"提示词: {prompt}")

    # 使用 predict 端点
    url = f"{API_BASE_URL}/v1beta/models/{model_id}:predict"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "instances": [{
            "prompt": prompt
        }],
        "parameters": {
            "sampleCount": 1,
            "aspectRatio": "1:1",  # 1:1, 16:9, 9:16, 4:3, 3:4
            "personGeneration": "allow_adult",
            "safetyFilterLevel": "block_few"
        }
    }

    print_info(f"请求 URL: {url}")

    try:
        start_time = datetime.now()
        response = requests.post(url, headers=headers, json=payload, timeout=60)
        elapsed = (datetime.now() - start_time).total_seconds()

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f} 秒")

        if response.status_code == 200:
            data = response.json()
            print_success("图片生成成功！")

            # 打印响应结构
            print(f"\n{Colors.BOLD}响应结构：{Colors.ENDC}")
            print(json.dumps(data, indent=2, ensure_ascii=False)[:2000])

            # 尝试提取和保存图片
            if "predictions" in data:
                predictions = data["predictions"]
                print_info(f"生成了 {len(predictions)} 张图片")

                for idx, pred in enumerate(predictions):
                    if "bytesBase64Encoded" in pred:
                        base64_data = pred["bytesBase64Encoded"]
                        image_bytes = base64.b64decode(base64_data)

                        # 保存图片
                        output_dir = Path("generated_images")
                        output_dir.mkdir(exist_ok=True)

                        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                        model_name = model_id.replace("/", "_").replace(":", "_")
                        filename = output_dir / f"{model_name}_{timestamp}_{idx+1}.png"

                        with open(filename, "wb") as f:
                            f.write(image_bytes)

                        print_success(f"图片已保存: {filename}")
                        print_info(f"图片大小: {len(image_bytes) / 1024:.2f} KB")

            return True

        else:
            print_error(f"生成失败 (HTTP {response.status_code})")
            try:
                error_data = response.json()
                print_error(f"错误: {json.dumps(error_data, indent=2, ensure_ascii=False)}")
            except:
                print_error(f"响应: {response.text[:500]}")
            return False

    except Exception as e:
        print_error(f"异常: {str(e)}")
        import traceback
        traceback.print_exc()
        return False

def test_gemini_image_gen(model_id: str, prompt: str):
    """测试 Gemini 图片生成"""
    print_info(f"模型: {model_id}")
    print_info(f"提示词: {prompt}")

    url = f"{API_BASE_URL}/v1beta/models/{model_id}:generateContent"

    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [{
            "parts": [{
                "text": f"Generate an image: {prompt}"
            }]
        }],
        "generationConfig": {
            "responseModalities": ["IMAGE"],
            "maxOutputTokens": 8192
        }
    }

    print_info(f"请求 URL: {url}")

    try:
        start_time = datetime.now()
        response = requests.post(url, headers=headers, json=payload, timeout=60)
        elapsed = (datetime.now() - start_time).total_seconds()

        print_info(f"HTTP 状态码: {response.status_code}")
        print_info(f"响应时间: {elapsed:.2f} 秒")

        if response.status_code == 200:
            data = response.json()
            print_success("请求成功！")

            # 打印响应
            print(f"\n{Colors.BOLD}响应内容：{Colors.ENDC}")
            print(json.dumps(data, indent=2, ensure_ascii=False)[:2000])

            # 尝试提取图片
            if "candidates" in data:
                candidates = data["candidates"]
                for idx, candidate in enumerate(candidates):
                    content = candidate.get("content", {})
                    parts = content.get("parts", [])

                    for part_idx, part in enumerate(parts):
                        # 检查是否有图片数据
                        if "inlineData" in part:
                            inline_data = part["inlineData"]
                            mime_type = inline_data.get("mimeType", "")
                            base64_data = inline_data.get("data", "")

                            if base64_data:
                                image_bytes = base64.b64decode(base64_data)

                                # 保存图片
                                output_dir = Path("generated_images")
                                output_dir.mkdir(exist_ok=True)

                                timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
                                ext = "png" if "png" in mime_type else "jpg"
                                filename = output_dir / f"gemini_image_{timestamp}_{idx+1}_{part_idx+1}.{ext}"

                                with open(filename, "wb") as f:
                                    f.write(image_bytes)

                                print_success(f"图片已保存: {filename}")
                                print_info(f"MIME: {mime_type}, 大小: {len(image_bytes) / 1024:.2f} KB")

            return True

        else:
            print_error(f"生成失败 (HTTP {response.status_code})")
            try:
                error_data = response.json()
                print_error(f"错误: {json.dumps(error_data, indent=2, ensure_ascii=False)}")
            except:
                print_error(f"响应: {response.text[:500]}")
            return False

    except Exception as e:
        print_error(f"异常: {str(e)}")
        import traceback
        traceback.print_exc()
        return False

def main():
    print_header("🎨 图片生成模型测试")

    prompt = "A serene mountain landscape at sunset with a lake reflecting the sky"

    # 测试 Imagen 4 Fast
    print_header("测试 Imagen 4 Fast")
    test_imagen_4("imagen-4.0-fast-generate-001", prompt)

    print("\n" + "="*70 + "\n")

    # 测试 Gemini 2.0 Flash Image Generation
    print_header("测试 Gemini 2.0 Flash Image Generation")
    test_gemini_image_gen("gemini-2.0-flash-exp-image-generation", prompt)

if __name__ == "__main__":
    main()
