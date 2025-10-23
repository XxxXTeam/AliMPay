# AliMPay 接入教程 / Integration Guide

本文档详细介绍如何将 AliMPay 支付系统集成到您的应用中，包含多种编程语言的示例代码。

This document provides detailed instructions on how to integrate AliMPay payment system into your application, with examples in multiple programming languages.

---

## 目录 / Table of Contents

- [接入前准备](#接入前准备--prerequisites)
- [获取商户信息](#获取商户信息--get-merchant-info)
- [签名算法](#签名算法--signature-algorithm)
- [创建支付订单](#创建支付订单--create-payment-order)
- [处理支付回调](#处理支付回调--handle-payment-callback)
- [查询订单状态](#查询订单状态--query-order-status)
- [完整示例代码](#完整示例代码--complete-examples)
- [测试指南](#测试指南--testing-guide)
- [常见问题](#常见问题--faq)

---

## 接入前准备 / Prerequisites

### 1. 确认服务已部署 / Ensure Service is Deployed

确保 AliMPay 服务已正确部署并可访问：

Ensure AliMPay service is properly deployed and accessible:

```bash
# 测试服务可用性 / Test service availability
curl http://your-domain.com/health

# 预期响应 / Expected response
# {"status":"ok","timestamp":"2024-01-15T12:00:00Z"}
```

### 2. 准备支付宝配置 / Prepare Alipay Configuration

确保已在配置文件中正确填写支付宝相关信息：

Ensure Alipay information is correctly filled in the configuration file:

- ✅ AppID (应用ID)
- ✅ 应用私钥 (Private Key)
- ✅ 支付宝公钥 (Alipay Public Key)
- ✅ 转账用户ID (Transfer User ID)
- ✅ 经营码图片 (Business QR Code Image)

### 3. 所需信息清单 / Required Information Checklist

集成前，您需要以下信息：

Before integration, you need the following information:

| 信息项 / Item | 说明 / Description | 示例 / Example |
|--------------|-------------------|---------------|
| API Base URL | 支付网关地址 / Payment gateway URL | `https://pay.example.com` |
| 商户ID / PID | 商户标识 / Merchant ID | `1001003549245339` |
| 商户密钥 / Key | 用于签名 / For signature | `abcdef1234567890` |
| 回调地址 / Notify URL | 异步通知地址 / Async notification URL | `https://your-site.com/notify` |
| 返回地址 / Return URL | 同步跳转地址 / Sync return URL | `https://your-site.com/success` |

---

## 获取商户信息 / Get Merchant Info

### 方式一：查看配置文件 / Method 1: Check Configuration File

首次运行后，系统会自动生成商户ID和密钥并保存在配置文件中：

After first run, the system will auto-generate merchant ID and key in configuration file:

```bash
# 查看配置文件
# View configuration file
cat configs/config.yaml | grep -A 2 "merchant:"
```

输出示例 / Output example:
```yaml
merchant:
  id: "1001003549245339"
  key: "f872e1c662d41cf218b5dfa8328ae455"
```

### 方式二：通过API查询 / Method 2: Query via API

```bash
# 查询商户信息
# Query merchant information
curl "http://your-domain.com/api?action=query&pid=YOUR_PID&key=YOUR_KEY"
```

响应示例 / Response example:
```json
{
  "code": 1,
  "pid": "1001003549245339",
  "key": "abcd****7890",
  "active": 1,
  "money": "0.00",
  "username": "Merchant",
  "rate": 96
}
```

---

## 签名算法 / Signature Algorithm

### 签名生成步骤 / Signature Generation Steps

AliMPay 使用 MD5 签名算法，与易支付/码支付完全兼容：

AliMPay uses MD5 signature algorithm, fully compatible with YiPay/CodePay:

**步骤 / Steps:**

1. **过滤参数** / Filter parameters
   - 移除空值参数 / Remove empty value parameters
   - 移除 `sign` 和 `sign_type` 参数 / Remove `sign` and `sign_type` parameters

2. **参数排序** / Sort parameters
   - 按参数名 ASCII 码升序排列 / Sort by parameter name in ASCII ascending order

3. **拼接字符串** / Concatenate string
   - 格式：`key1=value1&key2=value2&key3=value3`

4. **追加密钥** / Append key
   - 在字符串末尾拼接商户密钥 / Append merchant key at the end

5. **MD5 加密** / MD5 encryption
   - 对拼接后的字符串进行 MD5 加密并转为小写 / MD5 encrypt and convert to lowercase

### 签名示例 / Signature Example

**原始参数 / Original parameters:**
```
pid: 1001003549245339
type: alipay
out_trade_no: ORDER20240115001
name: 测试商品
money: 1.00
notify_url: http://example.com/notify
return_url: http://example.com/return
```

**商户密钥 / Merchant key:** `abcdef1234567890`

**排序后拼接 / After sorting and concatenation:**
```
money=1.00&name=测试商品&notify_url=http://example.com/notify&out_trade_no=ORDER20240115001&pid=1001003549245339&return_url=http://example.com/return&type=alipay
```

**加上密钥 / With key appended:**
```
money=1.00&name=测试商品&notify_url=http://example.com/notify&out_trade_no=ORDER20240115001&pid=1001003549245339&return_url=http://example.com/return&type=alipayabcdef1234567890
```

**MD5 签名 / MD5 signature:**
```
sign = md5(上述字符串).toLowerCase()
```

### 各语言签名实现 / Signature Implementation in Different Languages

#### PHP

```php
function generateSign($params, $key) {
    // 过滤空值和sign参数
    $filtered = array_filter($params, function($v, $k) {
        return $v !== '' && $k !== 'sign' && $k !== 'sign_type';
    }, ARRAY_FILTER_USE_BOTH);
    
    // 排序
    ksort($filtered);
    
    // 拼接
    $signStr = http_build_query($filtered);
    $signStr = urldecode($signStr);
    
    // 加密钥
    $signStr .= $key;
    
    // MD5并转小写
    return strtolower(md5($signStr));
}
```

#### Python

```python
import hashlib
from urllib.parse import urlencode

def generate_sign(params, key):
    # 过滤空值和sign参数
    filtered = {k: v for k, v in params.items() 
                if v and k not in ['sign', 'sign_type']}
    
    # 排序
    sorted_keys = sorted(filtered.keys())
    
    # 拼接
    sign_str = '&'.join([f'{k}={filtered[k]}' for k in sorted_keys])
    
    # 加密钥
    sign_str += key
    
    # MD5并转小写
    return hashlib.md5(sign_str.encode()).hexdigest().lower()
```

#### JavaScript (Node.js)

```javascript
const crypto = require('crypto');

function generateSign(params, key) {
    // 过滤空值和sign参数
    const filtered = Object.keys(params)
        .filter(k => params[k] && k !== 'sign' && k !== 'sign_type')
        .reduce((obj, k) => {
            obj[k] = params[k];
            return obj;
        }, {});
    
    // 排序
    const sortedKeys = Object.keys(filtered).sort();
    
    // 拼接
    const signStr = sortedKeys
        .map(k => `${k}=${filtered[k]}`)
        .join('&') + key;
    
    // MD5并转小写
    return crypto.createHash('md5').update(signStr).digest('hex').toLowerCase();
}
```

#### Java

```java
import java.security.MessageDigest;
import java.util.*;
import java.util.stream.Collectors;

public class SignatureUtil {
    public static String generateSign(Map<String, String> params, String key) {
        // 过滤空值和sign参数
        Map<String, String> filtered = params.entrySet().stream()
            .filter(e -> e.getValue() != null && !e.getValue().isEmpty())
            .filter(e -> !e.getKey().equals("sign") && !e.getKey().equals("sign_type"))
            .collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
        
        // 排序并拼接
        String signStr = filtered.entrySet().stream()
            .sorted(Map.Entry.comparingByKey())
            .map(e -> e.getKey() + "=" + e.getValue())
            .collect(Collectors.joining("&")) + key;
        
        // MD5并转小写
        return md5(signStr).toLowerCase();
    }
    
    private static String md5(String str) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] bytes = md.digest(str.getBytes("UTF-8"));
            StringBuilder sb = new StringBuilder();
            for (byte b : bytes) {
                sb.append(String.format("%02x", b));
            }
            return sb.toString();
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }
}
```

#### Go

```go
package main

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "sort"
    "strings"
)

func GenerateSign(params map[string]string, key string) string {
    // 过滤并排序
    var keys []string
    for k, v := range params {
        if v != "" && k != "sign" && k != "sign_type" {
            keys = append(keys, k)
        }
    }
    sort.Strings(keys)
    
    // 拼接
    var parts []string
    for _, k := range keys {
        parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
    }
    signStr := strings.Join(parts, "&") + key
    
    // MD5并转小写
    hash := md5.Sum([]byte(signStr))
    return strings.ToLower(hex.EncodeToString(hash[:]))
}
```

---

## 创建支付订单 / Create Payment Order

### 接口信息 / API Information

**接口地址 / Endpoint:**
- `/submit` (推荐 / Recommended)
- `/api/submit`
- `/submit.php` (易支付兼容 / YiPay compatible)

**请求方式 / Method:** `GET` / `POST`

**请求参数 / Request Parameters:**

| 参数名 / Parameter | 类型 / Type | 必填 / Required | 说明 / Description |
|-------------------|------------|----------------|-------------------|
| pid | string | 是 / Yes | 商户ID / Merchant ID |
| type | string | 是 / Yes | 支付方式，固定值：`alipay` / Payment type, fixed: `alipay` |
| out_trade_no | string | 是 / Yes | 商户订单号（唯一）/ Merchant order number (unique) |
| notify_url | string | 是 / Yes | 异步通知地址 / Async notification URL |
| return_url | string | 是 / Yes | 同步返回地址 / Sync return URL |
| name | string | 是 / Yes | 商品名称 / Product name |
| money | string | 是 / Yes | 订单金额（元）/ Order amount (yuan) |
| sitename | string | 否 / No | 网站名称 / Site name |
| sign | string | 是 / Yes | 签名 / Signature |
| sign_type | string | 否 / No | 签名类型，默认 MD5 / Signature type, default MD5 |

### 示例代码 / Example Code

#### PHP

```php
<?php
// 配置信息
$config = [
    'pid' => '1001003549245339',
    'key' => 'abcdef1234567890',
    'api_url' => 'https://pay.example.com'
];

// 订单信息
$orderData = [
    'pid' => $config['pid'],
    'type' => 'alipay',
    'out_trade_no' => 'ORDER' . time() . rand(1000, 9999),
    'notify_url' => 'https://your-site.com/notify.php',
    'return_url' => 'https://your-site.com/return.php',
    'name' => '测试商品',
    'money' => '0.01',
    'sitename' => '我的网站'
];

// 生成签名
$orderData['sign'] = generateSign($orderData, $config['key']);
$orderData['sign_type'] = 'MD5';

// 方式1: 直接跳转
$paymentUrl = $config['api_url'] . '/submit?' . http_build_query($orderData);
header('Location: ' . $paymentUrl);

// 方式2: 使用 cURL 获取支付信息
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $config['api_url'] . '/submit');
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($orderData));
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
$response = curl_exec($ch);
curl_close($ch);

$result = json_decode($response, true);
if ($result['code'] == 1) {
    // 成功，可以使用返回的信息
    echo '订单号: ' . $result['trade_no'] . '<br>';
    echo '支付金额: ' . $result['payment_amount'] . '<br>';
    echo '<img src="' . $result['qr_code'] . '" alt="支付二维码">';
} else {
    echo '创建订单失败: ' . $result['msg'];
}

function generateSign($params, $key) {
    $filtered = array_filter($params, function($v, $k) {
        return $v !== '' && $k !== 'sign' && $k !== 'sign_type';
    }, ARRAY_FILTER_USE_BOTH);
    ksort($filtered);
    $signStr = urldecode(http_build_query($filtered)) . $key;
    return strtolower(md5($signStr));
}
?>
```

#### Python

```python
import time
import random
import hashlib
import requests
from urllib.parse import urlencode

# 配置信息
CONFIG = {
    'pid': '1001003549245339',
    'key': 'abcdef1234567890',
    'api_url': 'https://pay.example.com'
}

def generate_sign(params, key):
    """生成签名"""
    filtered = {k: v for k, v in params.items() 
                if v and k not in ['sign', 'sign_type']}
    sorted_keys = sorted(filtered.keys())
    sign_str = '&'.join([f'{k}={filtered[k]}' for k in sorted_keys]) + key
    return hashlib.md5(sign_str.encode()).hexdigest().lower()

def create_payment(amount, product_name, order_no=None):
    """创建支付订单"""
    # 生成订单号
    if not order_no:
        order_no = f"ORDER{int(time.time())}{random.randint(1000, 9999)}"
    
    # 订单数据
    order_data = {
        'pid': CONFIG['pid'],
        'type': 'alipay',
        'out_trade_no': order_no,
        'notify_url': 'https://your-site.com/notify',
        'return_url': 'https://your-site.com/return',
        'name': product_name,
        'money': str(amount),
        'sitename': '我的网站'
    }
    
    # 生成签名
    order_data['sign'] = generate_sign(order_data, CONFIG['key'])
    order_data['sign_type'] = 'MD5'
    
    # 发起请求
    response = requests.post(f"{CONFIG['api_url']}/submit", data=order_data)
    result = response.json()
    
    if result['code'] == 1:
        print(f"订单创建成功!")
        print(f"系统订单号: {result['trade_no']}")
        print(f"商户订单号: {result['out_trade_no']}")
        print(f"支付金额: {result['payment_amount']}")
        print(f"支付页面: {result['payment_url']}")
        return result
    else:
        print(f"订单创建失败: {result['msg']}")
        return None

# 使用示例
if __name__ == '__main__':
    result = create_payment(0.01, '测试商品')
    if result:
        # 可以将用户重定向到支付页面
        payment_url = result['payment_url']
        print(f"\n请访问以下链接完成支付:\n{payment_url}")
```

#### JavaScript (Node.js)

```javascript
const crypto = require('crypto');
const axios = require('axios');

// 配置信息
const CONFIG = {
    pid: '1001003549245339',
    key: 'abcdef1234567890',
    apiUrl: 'https://pay.example.com'
};

// 生成签名
function generateSign(params, key) {
    const filtered = Object.keys(params)
        .filter(k => params[k] && k !== 'sign' && k !== 'sign_type')
        .reduce((obj, k) => {
            obj[k] = params[k];
            return obj;
        }, {});
    
    const sortedKeys = Object.keys(filtered).sort();
    const signStr = sortedKeys
        .map(k => `${k}=${filtered[k]}`)
        .join('&') + key;
    
    return crypto.createHash('md5').update(signStr).digest('hex').toLowerCase();
}

// 创建支付订单
async function createPayment(amount, productName, orderNo = null) {
    // 生成订单号
    if (!orderNo) {
        orderNo = `ORDER${Date.now()}${Math.floor(Math.random() * 9000) + 1000}`;
    }
    
    // 订单数据
    const orderData = {
        pid: CONFIG.pid,
        type: 'alipay',
        out_trade_no: orderNo,
        notify_url: 'https://your-site.com/notify',
        return_url: 'https://your-site.com/return',
        name: productName,
        money: amount.toString(),
        sitename: '我的网站'
    };
    
    // 生成签名
    orderData.sign = generateSign(orderData, CONFIG.key);
    orderData.sign_type = 'MD5';
    
    try {
        // 发起请求
        const response = await axios.post(`${CONFIG.apiUrl}/submit`, orderData);
        const result = response.data;
        
        if (result.code === 1) {
            console.log('订单创建成功!');
            console.log(`系统订单号: ${result.trade_no}`);
            console.log(`商户订单号: ${result.out_trade_no}`);
            console.log(`支付金额: ${result.payment_amount}`);
            console.log(`支付页面: ${result.payment_url}`);
            return result;
        } else {
            console.error(`订单创建失败: ${result.msg}`);
            return null;
        }
    } catch (error) {
        console.error('请求失败:', error.message);
        return null;
    }
}

// 使用示例
(async () => {
    const result = await createPayment(0.01, '测试商品');
    if (result) {
        console.log(`\n请访问以下链接完成支付:\n${result.payment_url}`);
    }
})();
```

---

## 处理支付回调 / Handle Payment Callback

支付成功后，AliMPay 会向您指定的 `notify_url` 发送 POST 请求。

After successful payment, AliMPay will send a POST request to your specified `notify_url`.

### 回调参数 / Callback Parameters

| 参数名 / Parameter | 类型 / Type | 说明 / Description |
|-------------------|------------|-------------------|
| pid | string | 商户ID / Merchant ID |
| trade_no | string | 系统订单号 / System order number |
| out_trade_no | string | 商户订单号 / Merchant order number |
| type | string | 支付方式 / Payment type |
| name | string | 商品名称 / Product name |
| money | string | 订单金额 / Order amount |
| trade_status | string | 交易状态：TRADE_SUCCESS |
| sign | string | 签名 / Signature |
| sign_type | string | 签名类型 / Signature type |

### 处理流程 / Processing Flow

1. **验证签名** / Verify signature
2. **检查订单状态** / Check order status
3. **处理业务逻辑** / Process business logic
4. **返回响应** / Return response

### 示例代码 / Example Code

#### PHP

```php
<?php
// notify.php

// 配置信息
$merchantKey = 'abcdef1234567890';

// 获取回调参数
$callbackData = $_POST;

// 1. 验证签名
$receivedSign = $callbackData['sign'];
unset($callbackData['sign']);
unset($callbackData['sign_type']);

$calculatedSign = generateSign($callbackData, $merchantKey);

if ($receivedSign !== $calculatedSign) {
    // 签名验证失败
    error_log('签名验证失败');
    exit('fail');
}

// 2. 获取订单信息
$tradeNo = $callbackData['trade_no'];
$outTradeNo = $callbackData['out_trade_no'];
$money = $callbackData['money'];
$tradeStatus = $callbackData['trade_status'];

// 3. 检查订单是否已处理
// 这里需要查询您的数据库
$order = getOrderByOutTradeNo($outTradeNo);

if (!$order) {
    error_log('订单不存在: ' . $outTradeNo);
    exit('fail');
}

if ($order['status'] == 'paid') {
    // 订单已处理，直接返回成功
    exit('success');
}

// 4. 验证金额
if ($order['amount'] != $money) {
    error_log('金额不匹配');
    exit('fail');
}

// 5. 处理订单（更新数据库等）
if ($tradeStatus == 'TRADE_SUCCESS') {
    // 更新订单状态为已支付
    updateOrderStatus($outTradeNo, 'paid', $tradeNo);
    
    // 执行业务逻辑（发货、开通服务等）
    processOrderBusiness($outTradeNo);
    
    // 记录日志
    error_log('订单支付成功: ' . $outTradeNo);
    
    // 返回成功
    exit('success');
} else {
    error_log('交易状态异常: ' . $tradeStatus);
    exit('fail');
}

function generateSign($params, $key) {
    $filtered = array_filter($params, function($v, $k) {
        return $v !== '';
    }, ARRAY_FILTER_USE_BOTH);
    ksort($filtered);
    $signStr = urldecode(http_build_query($filtered)) . $key;
    return strtolower(md5($signStr));
}

function getOrderByOutTradeNo($outTradeNo) {
    // 从数据库查询订单
    // Query order from database
    // 返回订单信息或 null
    // Return order info or null
}

function updateOrderStatus($outTradeNo, $status, $tradeNo) {
    // 更新订单状态到数据库
    // Update order status in database
}

function processOrderBusiness($outTradeNo) {
    // 执行业务逻辑
    // Execute business logic
}
?>
```

#### Python (Flask)

```python
from flask import Flask, request
import hashlib

app = Flask(__name__)

# 配置信息
MERCHANT_KEY = 'abcdef1234567890'

def generate_sign(params, key):
    """生成签名"""
    filtered = {k: v for k, v in params.items() if v}
    sorted_keys = sorted(filtered.keys())
    sign_str = '&'.join([f'{k}={filtered[k]}' for k in sorted_keys]) + key
    return hashlib.md5(sign_str.encode()).hexdigest().lower()

@app.route('/notify', methods=['POST'])
def payment_notify():
    """处理支付回调"""
    callback_data = request.form.to_dict()
    
    # 1. 验证签名
    received_sign = callback_data.pop('sign', '')
    callback_data.pop('sign_type', '')
    
    calculated_sign = generate_sign(callback_data, MERCHANT_KEY)
    
    if received_sign != calculated_sign:
        app.logger.error('签名验证失败')
        return 'fail'
    
    # 2. 获取订单信息
    trade_no = callback_data.get('trade_no')
    out_trade_no = callback_data.get('out_trade_no')
    money = callback_data.get('money')
    trade_status = callback_data.get('trade_status')
    
    # 3. 检查订单是否已处理
    order = get_order_by_out_trade_no(out_trade_no)
    
    if not order:
        app.logger.error(f'订单不存在: {out_trade_no}')
        return 'fail'
    
    if order['status'] == 'paid':
        # 订单已处理
        return 'success'
    
    # 4. 验证金额
    if str(order['amount']) != money:
        app.logger.error('金额不匹配')
        return 'fail'
    
    # 5. 处理订单
    if trade_status == 'TRADE_SUCCESS':
        # 更新订单状态
        update_order_status(out_trade_no, 'paid', trade_no)
        
        # 执行业务逻辑
        process_order_business(out_trade_no)
        
        app.logger.info(f'订单支付成功: {out_trade_no}')
        return 'success'
    else:
        app.logger.error(f'交易状态异常: {trade_status}')
        return 'fail'

def get_order_by_out_trade_no(out_trade_no):
    """从数据库查询订单"""
    # 实现数据库查询逻辑
    pass

def update_order_status(out_trade_no, status, trade_no):
    """更新订单状态"""
    # 实现数据库更新逻辑
    pass

def process_order_business(out_trade_no):
    """执行业务逻辑"""
    # 实现业务逻辑
    pass

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
```

#### JavaScript (Express)

```javascript
const express = require('express');
const crypto = require('crypto');
const bodyParser = require('body-parser');

const app = express();
app.use(bodyParser.urlencoded({ extended: true }));

// 配置信息
const MERCHANT_KEY = 'abcdef1234567890';

// 生成签名
function generateSign(params, key) {
    const filtered = Object.keys(params)
        .filter(k => params[k])
        .reduce((obj, k) => {
            obj[k] = params[k];
            return obj;
        }, {});
    
    const sortedKeys = Object.keys(filtered).sort();
    const signStr = sortedKeys
        .map(k => `${k}=${filtered[k]}`)
        .join('&') + key;
    
    return crypto.createHash('md5').update(signStr).digest('hex').toLowerCase();
}

// 支付回调处理
app.post('/notify', async (req, res) => {
    const callbackData = { ...req.body };
    
    // 1. 验证签名
    const receivedSign = callbackData.sign;
    delete callbackData.sign;
    delete callbackData.sign_type;
    
    const calculatedSign = generateSign(callbackData, MERCHANT_KEY);
    
    if (receivedSign !== calculatedSign) {
        console.error('签名验证失败');
        return res.send('fail');
    }
    
    // 2. 获取订单信息
    const { trade_no, out_trade_no, money, trade_status } = req.body;
    
    // 3. 检查订单是否已处理
    const order = await getOrderByOutTradeNo(out_trade_no);
    
    if (!order) {
        console.error(`订单不存在: ${out_trade_no}`);
        return res.send('fail');
    }
    
    if (order.status === 'paid') {
        // 订单已处理
        return res.send('success');
    }
    
    // 4. 验证金额
    if (order.amount.toString() !== money) {
        console.error('金额不匹配');
        return res.send('fail');
    }
    
    // 5. 处理订单
    if (trade_status === 'TRADE_SUCCESS') {
        // 更新订单状态
        await updateOrderStatus(out_trade_no, 'paid', trade_no);
        
        // 执行业务逻辑
        await processOrderBusiness(out_trade_no);
        
        console.log(`订单支付成功: ${out_trade_no}`);
        return res.send('success');
    } else {
        console.error(`交易状态异常: ${trade_status}`);
        return res.send('fail');
    }
});

async function getOrderByOutTradeNo(outTradeNo) {
    // 从数据库查询订单
    // Query order from database
}

async function updateOrderStatus(outTradeNo, status, tradeNo) {
    // 更新订单状态
    // Update order status
}

async function processOrderBusiness(outTradeNo) {
    // 执行业务逻辑
    // Execute business logic
}

app.listen(3000, () => {
    console.log('Callback server running on port 3000');
});
```

---

## 查询订单状态 / Query Order Status

### 接口信息 / API Information

**接口地址 / Endpoint:**
- `/api/order`
- `/mapi?act=order` (兼容接口 / Compatible endpoint)

**请求方式 / Method:** `GET` / `POST`

**请求参数 / Request Parameters:**

| 参数名 / Parameter | 类型 / Type | 必填 / Required | 说明 / Description |
|-------------------|------------|----------------|-------------------|
| pid | string | 是 / Yes | 商户ID / Merchant ID |
| out_trade_no | string | 是 / Yes | 商户订单号 / Merchant order number |

### 示例代码 / Example Code

#### PHP

```php
<?php
function queryOrder($outTradeNo) {
    $config = [
        'pid' => '1001003549245339',
        'api_url' => 'https://pay.example.com'
    ];
    
    $params = [
        'pid' => $config['pid'],
        'out_trade_no' => $outTradeNo
    ];
    
    $url = $config['api_url'] . '/api/order?' . http_build_query($params);
    $response = file_get_contents($url);
    return json_decode($response, true);
}

$result = queryOrder('ORDER20240115001');
print_r($result);
?>
```

#### Python

```python
import requests

def query_order(out_trade_no):
    config = {
        'pid': '1001003549245339',
        'api_url': 'https://pay.example.com'
    }
    
    params = {
        'pid': config['pid'],
        'out_trade_no': out_trade_no
    }
    
    response = requests.get(f"{config['api_url']}/api/order", params=params)
    return response.json()

result = query_order('ORDER20240115001')
print(result)
```

---

## 完整示例代码 / Complete Examples

完整的示例代码已包含在项目仓库的 `examples` 目录中（即将添加）：

Complete example code is included in the project repository's `examples` directory (coming soon):

- `examples/php/` - PHP 示例 / PHP examples
- `examples/python/` - Python 示例 / Python examples
- `examples/nodejs/` - Node.js 示例 / Node.js examples
- `examples/java/` - Java 示例 / Java examples
- `examples/go/` - Go 示例 / Go examples

---

## 测试指南 / Testing Guide

### 1. 使用测试脚本 / Using Test Scripts

项目提供了测试脚本用于快速测试：

The project provides test scripts for quick testing:

```bash
# Python 测试脚本
python3 test_payment.py

# 或使用项目提供的脚本
python3 generate_payment_url.py
python3 generate_payment_url_v2.py
```

### 2. 手动测试流程 / Manual Testing Process

**步骤 / Steps:**

1. **创建测试订单** / Create test order
   ```bash
   curl -X POST "http://localhost:8080/submit" \
     -d "pid=YOUR_PID" \
     -d "type=alipay" \
     -d "out_trade_no=TEST$(date +%s)" \
     -d "name=测试商品" \
     -d "money=0.01" \
     -d "notify_url=http://example.com/notify" \
     -d "return_url=http://example.com/return" \
     -d "sign=YOUR_SIGN"
   ```

2. **扫码支付** / Scan and pay
   - 使用支付宝扫描返回的二维码
   - 支付测试金额（0.01元）

3. **验证回调** / Verify callback
   - 检查 notify_url 是否收到回调
   - 验证回调数据的签名

4. **查询订单** / Query order
   ```bash
   curl "http://localhost:8080/api/order?pid=YOUR_PID&out_trade_no=TEST123"
   ```

### 3. 测试注意事项 / Testing Notes

- ✅ 使用小金额测试（0.01元）
- ✅ 确保回调地址可公网访问
- ✅ 检查签名验证逻辑
- ✅ 测试订单重复支付情况
- ✅ 测试订单超时情况

---

## 常见问题 / FAQ

### Q1: 如何获取商户ID和密钥？

**A:** 首次启动服务后，系统会自动生成并保存在配置文件 `configs/config.yaml` 的 `merchant` 部分。

### Q2: 签名验证总是失败？

**A:** 请检查：
1. 参数是否按 ASCII 码排序
2. 是否正确过滤了 `sign` 和 `sign_type` 参数
3. URL 编码处理是否正确
4. 商户密钥是否正确
5. MD5 是否转为小写

### Q3: 回调地址收不到通知？

**A:** 请确认：
1. 回调地址必须是公网可访问的 HTTP/HTTPS 地址
2. 服务器防火墙是否开放
3. 回调接口是否返回 `success`
4. 查看 AliMPay 日志了解详细错误

### Q4: 经营码和转账模式如何选择？

**A:**
- **经营码模式**（推荐）：使用固定二维码，系统通过金额匹配订单，到账快
- **转账模式**：每个订单生成独立二维码，更灵活但需要额外配置

### Q5: 测试环境如何配置？

**A:** 可以使用支付宝沙箱环境：
1. 修改配置文件中的 `server_url` 为沙箱网关
2. 使用沙箱应用的 AppID 和密钥
3. 下载沙箱版支付宝 APP 进行测试

---

## 技术支持 / Technical Support

如需帮助，请通过以下方式联系：

For assistance, please contact via:

- **GitHub Issues**: https://github.com/chanhanzhan/AliMPay/issues
- **文档 / Documentation**: https://github.com/chanhanzhan/AliMPay/tree/main/docs
- **API文档 / API Docs**: https://github.com/chanhanzhan/AliMPay/blob/main/docs/API.md
- **Email**: support@openel.top

---

**祝您接入顺利！/ Happy Integrating!** 🎉
