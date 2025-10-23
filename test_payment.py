#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""快速测试支付URL生成"""

import hashlib
import urllib.parse

def md5(text):
    return hashlib.md5(text.encode('utf-8')).hexdigest()

def generate_sign(params, key):
    # 移除sign和sign_type，移除空值
    filtered = {k: v for k, v in params.items() 
                if v and k not in ['sign', 'sign_type']}
    
    # 按键名排序
    sorted_keys = sorted(filtered.keys())
    
    # 构建签名字符串
    sign_str = '&'.join([f'{k}={filtered[k]}' for k in sorted_keys])
    
    # 加上密钥并计算MD5
    sign_str_with_key = sign_str + key
    
    print(f"\n签名计算过程:")
    print(f"参数: {filtered}")
    print(f"签名串: {sign_str}")
    print(f"加密钥: {sign_str_with_key}")
    
    signature = md5(sign_str_with_key)
    print(f"签  名: {signature}\n")
    
    return signature

# 商户信息
PID = "1001001276912812"
KEY = "f872e1c662d41cf218b5dfa8328ae455"
BASE_URL = "http://localhost:8080"

# 支付参数
params = {
    'pid': PID,
    'type': 'alipay',
    'out_trade_no': 'TEST' + str(int(__import__('time').time())),
    'notify_url': 'http://example.com/notify',
    'return_url': 'http://example.com/return',
    'name': '测试商品',
    'money': '0.01',
    'sign_type': 'MD5'
}

# 生成签名
sign = generate_sign(params, KEY)
params['sign'] = sign

# 构建URL
query_string = urllib.parse.urlencode(params)
full_url = f"{BASE_URL}/submit?{query_string}"

print("="*70)
print("支付URL:")
print(full_url)
print("="*70)
print("\n复制此URL到浏览器测试\n")

