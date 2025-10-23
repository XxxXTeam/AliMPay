#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
支付宝深链接功能测试脚本
Test script for Alipay Deep Link feature
"""

import requests
import json
from urllib.parse import urlencode

# 配置
BASE_URL = "http://localhost:8080"
QR_CODE_ID = "fkx12345678901234"  # 示例二维码ID，请替换为真实ID
AMOUNT = 1.23
REMARK = "测试订单"

def print_header(title):
    """打印测试标题"""
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60)

def print_test(test_name, url, response):
    """打印测试结果"""
    print(f"\n{test_name}")
    print("-" * 60)
    print(f"请求: {url}")
    print(f"状态码: {response.status_code}")
    print("响应:")
    try:
        print(json.dumps(response.json(), indent=2, ensure_ascii=False))
    except:
        print(response.text)

def test_default_qr_code_id():
    """测试1: 使用默认二维码ID"""
    params = {
        "amount": AMOUNT,
        "remark": REMARK
    }
    url = f"{BASE_URL}/alipay/link?{urlencode(params)}"
    response = requests.get(url)
    print_test("测试1: 使用默认二维码ID生成深链接", url, response)
    return response

def test_custom_qr_code_id():
    """测试2: 使用自定义二维码ID"""
    params = {
        "qr_code_id": QR_CODE_ID,
        "amount": AMOUNT,
        "remark": REMARK
    }
    url = f"{BASE_URL}/alipay/link?{urlencode(params)}"
    response = requests.get(url)
    print_test("测试2: 使用自定义二维码ID生成深链接", url, response)
    return response

def test_no_amount():
    """测试3: 不带金额和备注"""
    params = {
        "qr_code_id": QR_CODE_ID
    }
    url = f"{BASE_URL}/alipay/link?{urlencode(params)}"
    response = requests.get(url)
    print_test("测试3: 仅生成二维码链接（不带金额）", url, response)
    return response

def test_missing_qr_code_id():
    """测试4: 错误场景 - 缺少二维码ID"""
    params = {
        "amount": AMOUNT
    }
    url = f"{BASE_URL}/alipay/link?{urlencode(params)}"
    response = requests.get(url)
    print_test("测试4: 错误场景 - 缺少二维码ID", url, response)
    return response

def test_invalid_amount():
    """测试5: 错误场景 - 无效金额"""
    params = {
        "qr_code_id": QR_CODE_ID,
        "amount": "invalid"
    }
    url = f"{BASE_URL}/alipay/link?{urlencode(params)}"
    response = requests.get(url)
    print_test("测试5: 错误场景 - 无效金额格式", url, response)
    return response

def test_redirect_to_alipay():
    """测试6: 直接重定向到支付宝"""
    params = {
        "qr_code_id": QR_CODE_ID,
        "amount": AMOUNT,
        "remark": REMARK
    }
    url = f"{BASE_URL}/alipay/pay?{urlencode(params)}"
    response = requests.get(url, allow_redirects=False)
    print(f"\n测试6: 直接重定向到支付宝")
    print("-" * 60)
    print(f"请求: {url}")
    print(f"状态码: {response.status_code}")
    if "Location" in response.headers:
        print(f"重定向URL: {response.headers['Location']}")
    else:
        print("未发现重定向")
    return response

def main():
    """主函数"""
    print_header("支付宝深链接功能测试脚本")
    
    print("\n配置信息:")
    print(f"  服务地址: {BASE_URL}")
    print(f"  二维码ID: {QR_CODE_ID}")
    print(f"  支付金额: ¥{AMOUNT}")
    print(f"  备注信息: {REMARK}")
    
    try:
        # 运行所有测试
        test_default_qr_code_id()
        test_custom_qr_code_id()
        test_no_amount()
        test_missing_qr_code_id()
        test_invalid_amount()
        test_redirect_to_alipay()
        
        print_header("测试完成")
        print("\n提示：")
        print("1. 请确保服务已启动: make run")
        print("2. 深链接仅在移动设备上有效")
        print("3. 需要在配置文件中设置正确的 qr_code_id")
        print()
        
    except requests.exceptions.ConnectionError:
        print("\n❌ 错误: 无法连接到服务器")
        print(f"   请确保服务已在 {BASE_URL} 启动")
        print("   运行命令: make run")
        print()
    except Exception as e:
        print(f"\n❌ 错误: {e}")
        print()

if __name__ == "__main__":
    main()
