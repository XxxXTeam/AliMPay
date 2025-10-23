#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
支付URL生成工具 - 增强版
用于生成带有正确签名的支付链接
"""

import hashlib
import urllib.parse
import sys

def md5(text):
    """计算MD5哈希"""
    return hashlib.md5(text.encode('utf-8')).hexdigest()

def generate_sign(params, key):
    """
    生成签名
    :param params: 参数字典
    :param key: 商户密钥
    :return: 签名字符串
    """
    # 移除sign和sign_type，移除空值
    filtered = {k: v for k, v in params.items() 
                if v and k not in ['sign', 'sign_type']}
    
    # 按键名排序
    sorted_keys = sorted(filtered.keys())
    
    # 构建签名字符串
    sign_str = '&'.join([f'{k}={filtered[k]}' for k in sorted_keys])
    
    # 加上密钥并计算MD5
    sign_str_with_key = sign_str + key
    
    print(f"\n[签名计算过程]")
    print(f"1. 参与签名的参数: {filtered}")
    print(f"2. 排序后的键: {sorted_keys}")
    print(f"3. 签名字符串: {sign_str}")
    print(f"4. 加上密钥: {sign_str_with_key}")
    
    signature = md5(sign_str_with_key)
    print(f"5. 最终签名: {signature}")
    
    return signature

def generate_payment_url(base_url, pid, merchant_key, **kwargs):
    """
    生成支付URL
    :param base_url: 基础URL（如: http://localhost:8080）
    :param pid: 商户ID
    :param merchant_key: 商户密钥
    :param kwargs: 其他参数
    :return: 完整的支付URL
    """
    # 默认参数
    params = {
        'pid': pid,
        'type': kwargs.get('type', 'alipay'),
        'out_trade_no': kwargs.get('out_trade_no', 'TEST' + str(int(__import__('time').time()))),
        'notify_url': kwargs.get('notify_url', 'http://example.com/notify'),
        'return_url': kwargs.get('return_url', 'http://example.com/return'),
        'name': kwargs.get('name', '测试商品'),
        'money': str(kwargs.get('money', '0.01')),
        'sign_type': 'MD5'
    }
    
    # 可选参数
    if 'sitename' in kwargs and kwargs['sitename']:
        params['sitename'] = kwargs['sitename']
    
    # 生成签名
    sign = generate_sign(params, merchant_key)
    params['sign'] = sign
    
    # 构建URL
    query_string = urllib.parse.urlencode(params)
    full_url = f"{base_url}/submit?{query_string}"
    
    return full_url, params

def main():
    """主函数"""
    print("=" * 70)
    print("支付URL生成工具 - AliMPay")
    print("=" * 70)
    
    # 配置参数（从config.yaml中获取）
    BASE_URL = "http://localhost:8080"  # 修改为您的实际地址
    
    # 商户信息（请确保与configs/config.yaml中的一致）
    print("\n请输入商户信息（留空使用默认值）:")
    
    pid_input = input("商户ID [1001001276912812]: ").strip()
    PID = pid_input if pid_input else "1001001276912812"
    
    key_input = input("商户密钥 [f872e1c662d41cf218b5dfa8328ae455]: ").strip()
    MERCHANT_KEY = key_input if key_input else "f872e1c662d41cf218b5dfa8328ae455"
    
    # 支付参数
    print("\n请输入支付参数（留空使用默认值）:")
    
    money_input = input("支付金额 [0.01]: ").strip()
    money = money_input if money_input else "0.01"
    
    name_input = input("商品名称 [测试商品]: ").strip()
    name = name_input if name_input else "测试商品"
    
    out_trade_no_input = input("商户订单号 [自动生成]: ").strip()
    out_trade_no = out_trade_no_input if out_trade_no_input else f"TEST{int(__import__('time').time())}"
    
    # 生成支付URL
    url, params = generate_payment_url(
        base_url=BASE_URL,
        pid=PID,
        merchant_key=MERCHANT_KEY,
        money=money,
        name=name,
        out_trade_no=out_trade_no
    )
    
    print("\n" + "=" * 70)
    print("生成结果")
    print("=" * 70)
    print(f"\n完整URL:")
    print(url)
    
    print(f"\n所有参数:")
    for k, v in sorted(params.items()):
        print(f"  {k}: {v}")
    
    print("\n" + "=" * 70)
    print("提示:")
    print("1. 请确保商户ID和密钥与configs/config.yaml中的一致")
    print("2. 可以直接在浏览器中访问上面的URL进行测试")
    print("3. 如果签名验证失败，请检查商户密钥是否正确")
    print("=" * 70)
    
    # 生成curl命令
    print(f"\n使用curl测试:")
    print(f'curl "{url}"')

if __name__ == "__main__":
    main()

