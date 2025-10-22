#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
快速生成支付页面URL
用于在浏览器中打开支付页面进行测试
"""

import hashlib
import time
from urllib.parse import urlencode, quote

# 配置（从 test_example.sh 中读取的配置）
PID = "1001003549245339"
KEY = "cd5dcdcbef4da67b9f3a01b1e391ab86"
BASE_URL = "http://localhost:8080/submit"

def generate_sign(params, key):
    """生成MD5签名"""
    sorted_params = sorted(params.items())
    sign_str = '&'.join([f"{k}={v}" for k, v in sorted_params]) + key
    return hashlib.md5(sign_str.encode()).hexdigest()

def generate_payment_url(amount="0.01", product_name="测试商品"):
    """生成支付URL"""
    # 订单参数
    params = {
        'pid': PID,
        'type': 'alipay',
        'out_trade_no': f'TEST{int(time.time())}',
        'notify_url': 'http://example.com/notify',
        'return_url': 'http://example.com/return',
        'name': product_name,
        'money': amount,
    }

    # 生成签名
    sign = generate_sign(params, KEY)
    
    # 添加签名
    params['sign'] = sign
    params['sign_type'] = 'MD5'

    # 生成URL
    url = f"{BASE_URL}?{urlencode(params)}"
    
    return url, params

def main():
    print("\n" + "="*70)
    print(" 💳 支付页面URL生成器")
    print("="*70)
    
    # 获取用户输入（可选）
    print("\n请输入支付信息（直接回车使用默认值）：")
    amount = input("支付金额 [0.01]: ").strip() or "0.01"
    product_name = input("商品名称 [测试商品]: ").strip() or "测试商品"
    
    # 生成URL
    url, params = generate_payment_url(amount, product_name)
    
    # 显示结果
    print("\n" + "-"*70)
    print("📦 订单信息:")
    print("-"*70)
    print(f"  商户ID:     {params['pid']}")
    print(f"  订单号:     {params['out_trade_no']}")
    print(f"  商品名称:   {params['name']}")
    print(f"  支付金额:   ¥{params['money']}")
    print(f"  签名:       {params['sign']}")
    
    print("\n" + "-"*70)
    print("🔗 支付页面URL:")
    print("-"*70)
    print(url)
    
    print("\n" + "-"*70)
    print("📱 使用方法:")
    print("-"*70)
    print("  1. 确保服务已启动: make run")
    print("  2. 复制上面的URL到浏览器打开")
    print("  3. 或者使用命令自动打开:")
    print(f"     xdg-open '{url}'")
    print("-"*70)
    
    # 尝试自动打开浏览器
    try_open = input("\n是否自动在浏览器中打开？(y/n) [y]: ").strip().lower()
    if try_open in ('', 'y', 'yes'):
        import webbrowser
        try:
            webbrowser.open(url)
            print("✅ 已在浏览器中打开支付页面！")
        except Exception as e:
            print(f"❌ 自动打开失败: {e}")
            print("请手动复制URL到浏览器打开")
    
    print("\n✨ 完成！\n")

if __name__ == '__main__':
    main()

