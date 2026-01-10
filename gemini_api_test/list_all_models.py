#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
列出所有可用的 Gemini 模型
"""

import requests
import json

API_BASE_URL = "https://router.aitokencloud.com"
API_KEY = "sk-55d53fa110deaf67d76d051888348cca81501d8b1575735c89c3143a434d0b01"

def list_models():
    """获取所有可用模型"""
    url = f"{API_BASE_URL}/v1beta/models"
    headers = {
        "Authorization": f"Bearer {API_KEY}"
    }

    try:
        response = requests.get(url, headers=headers, timeout=30)

        if response.status_code == 200:
            data = response.json()
            models = data.get("models", [])

            print(f"共发现 {len(models)} 个可用模型：\n")
            print("=" * 80)

            # 按名称分组
            gemini_models = []

            for model in models:
                name = model.get("name", "")
                display_name = model.get("displayName", "")
                description = model.get("description", "")
                input_limit = model.get("inputTokenLimit", 0)
                output_limit = model.get("outputTokenLimit", 0)
                version = model.get("version", "")

                model_info = {
                    "name": name,
                    "display_name": display_name,
                    "description": description,
                    "input_limit": input_limit,
                    "output_limit": output_limit,
                    "version": version
                }
                gemini_models.append(model_info)

            # 排序并打印
            gemini_models.sort(key=lambda x: x["name"])

            for idx, model in enumerate(gemini_models, 1):
                print(f"\n{idx}. {model['name']}")
                if model['display_name']:
                    print(f"   显示名称: {model['display_name']}")
                if model['version']:
                    print(f"   版本: {model['version']}")
                if model['description']:
                    print(f"   描述: {model['description']}")
                print(f"   输入限制: {model['input_limit']:,} tokens")
                print(f"   输出限制: {model['output_limit']:,} tokens")

            print("\n" + "=" * 80)

            # 分类统计
            gemini_1_count = sum(1 for m in gemini_models if "gemini-1" in m["name"].lower())
            gemini_2_count = sum(1 for m in gemini_models if "gemini-2" in m["name"].lower())
            gemini_3_count = sum(1 for m in gemini_models if "gemini-3" in m["name"].lower())

            print("\n📊 模型版本统计：")
            print(f"  - Gemini 1.x 系列: {gemini_1_count} 个模型")
            print(f"  - Gemini 2.x 系列: {gemini_2_count} 个模型")
            print(f"  - Gemini 3.x 系列: {gemini_3_count} 个模型")

        else:
            print(f"❌ 请求失败: HTTP {response.status_code}")
            print(response.text)

    except Exception as e:
        print(f"❌ 发生错误: {str(e)}")

if __name__ == "__main__":
    list_models()
