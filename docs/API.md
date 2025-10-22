# AliMPay API 文档

## 目录

- [接口说明](#接口说明)
- [签名算法](#签名算法)
- [支付接口](#支付接口)
- [查询接口](#查询接口)
- [管理接口](#管理接口)
- [错误码](#错误码)
- [示例代码](#示例代码)

---

## 接口说明

### 基本信息

- **接口协议**: HTTP/HTTPS
- **请求方式**: GET / POST
- **响应格式**: JSON
- **字符编码**: UTF-8
- **签名算法**: MD5

### 通用参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| sign | string | 是 | 签名 |
| sign_type | string | 否 | 签名类型，默认MD5 |

### 通用响应格式

```json
{
  "code": 1,           // 状态码：1=成功，-1=失败
  "msg": "SUCCESS",    // 消息
  "data": {}           // 数据
}
```

---

## 签名算法

### MD5 签名步骤

1. **参数排序**: 将所有请求参数（除 `sign` 和 `sign_type`）按参数名ASCII码升序排列
2. **拼接字符串**: 格式为 `key1=value1&key2=value2&key3=value3`
3. **追加密钥**: 在字符串末尾追加商户密钥 `{merchant_key}`
4. **MD5加密**: 对拼接后的字符串进行MD5加密，转小写

### 签名示例

**原始参数**:
```
pid=1001003549245339
type=alipay
out_trade_no=TEST20240115001
name=测试商品
money=1.00
notify_url=http://example.com/notify
return_url=http://example.com/return
```

**商户密钥**: `abcdef1234567890`

**排序后拼接**:
```
money=1.00&name=测试商品&notify_url=http://example.com/notify&out_trade_no=TEST20240115001&pid=1001003549245339&return_url=http://example.com/return&type=alipayabcdef1234567890
```

**签名结果**:
```
sign=md5({排序拼接字符串}) // 转小写
```

---

## 支付接口

### 1. 创建支付订单

创建新的支付订单并返回支付二维码。

**接口地址**: 

- `/submit` (GET/POST)
- `/api/submit` (GET/POST)
- `/submit.php` (兼容接口)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| type | string | 是 | 支付方式，固定值：alipay |
| out_trade_no | string | 是 | 商户订单号，唯一标识 |
| notify_url | string | 是 | 异步通知地址 |
| return_url | string | 是 | 同步返回地址 |
| name | string | 是 | 商品名称 |
| money | string | 是 | 订单金额，精确到分 |
| sitename | string | 否 | 网站名称 |
| sign | string | 是 | 签名 |
| sign_type | string | 否 | 签名类型，默认MD5 |

**响应示例**:

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "pid": "1001003549245339",
  "trade_no": "20240115120000123456",
  "out_trade_no": "TEST20240115001",
  "money": "1.00",
  "payment_amount": 1.01,
  "create_time": "2024-01-15 12:00:00",
  "payment_url": "http://your-domain.com/pay?trade_no=xxx&amount=1.01",
  "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEU...",
  "business_qr_mode": true,
  "payment_tips": [
    "请务必支付准确金额：1.01 元",
    "支付时无需填写备注信息",
    "请在5分钟内完成支付"
  ]
}
```

**字段说明**:

- `trade_no`: 系统订单号
- `payment_amount`: 实际支付金额（经营码模式可能与订单金额不同）
- `payment_url`: 支付页面URL
- `qr_code`: Base64编码的二维码图片
- `business_qr_mode`: 是否为经营码模式

### 2. 异步通知

支付成功后，系统会向 `notify_url` 发送POST通知。

**通知参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| pid | string | 商户ID |
| trade_no | string | 系统订单号 |
| out_trade_no | string | 商户订单号 |
| type | string | 支付方式 |
| name | string | 商品名称 |
| money | string | 订单金额 |
| trade_status | string | 交易状态：TRADE_SUCCESS |
| sign | string | 签名 |
| sign_type | string | 签名类型 |

**响应要求**:

商户必须返回字符串 `success` 或 `ok` 表示接收成功，否则系统会重试通知。

---

## 查询接口

### 1. 查询订单状态

**接口地址**: 

- `/api/order` (GET/POST)
- `/mapi?act=order` (兼容接口)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| out_trade_no | string | 是 | 商户订单号 |
| key | string | 否 | 商户密钥（可选） |

**响应示例**:

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "trade_no": "20240115120000123456",
  "out_trade_no": "TEST20240115001",
  "type": "alipay",
  "pid": "1001003549245339",
  "name": "测试商品",
  "money": "1.00",
  "status": 1,
  "addtime": "2024-01-15 12:00:00",
  "endtime": "2024-01-15 12:01:30"
}
```

**状态说明**:

- `0`: 待支付
- `1`: 已支付
- `2`: 已关闭

### 2. 查询订单列表

**接口地址**: `/api/orders` (GET/POST)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |
| limit | int | 否 | 返回数量，默认20 |

**响应示例**:

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "count": 2,
  "orders": [
    {
      "trade_no": "20240115120000123456",
      "out_trade_no": "TEST20240115001",
      "type": "alipay",
      "name": "测试商品",
      "money": "1.00",
      "status": 1,
      "addtime": "2024-01-15 12:00:00",
      "endtime": "2024-01-15 12:01:30"
    }
  ]
}
```

### 3. 查询商户信息

**接口地址**: 

- `/api?action=query` (GET/POST)
- `/api/query` (GET/POST)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |

**响应示例**:

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

## 管理接口

### 1. 标记订单已支付

手动标记订单为已支付状态。

**接口地址**: `/admin?action=mark_paid` (POST)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |
| out_trade_no | string | 是 | 商户订单号 |

**响应示例**:

```json
{
  "success": true,
  "message": "Order marked as paid successfully",
  "order": {
    "trade_no": "20240115120000123456",
    "out_trade_no": "TEST20240115001",
    "status": "paid",
    "pay_time": "2024-01-15 12:05:00",
    "payment_amount": 1.01
  },
  "notification": {
    "sent": true,
    "url": "http://example.com/notify"
  }
}
```

### 2. 取消订单

**接口地址**: `/admin?action=cancel` (POST)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |
| trade_no | string | 是 | 系统订单号 |

**响应示例**:

```json
{
  "success": true,
  "message": "Order cancelled successfully",
  "order": {
    "trade_no": "20240115120000123456",
    "status": "closed"
  }
}
```

### 3. 获取订单列表

**接口地址**: `/admin/orders` (GET)

**响应示例**:

```json
{
  "code": 1,
  "msg": "success",
  "orders": [...]
}
```

### 4. 关闭订单

**接口地址**: `/api/close` (GET/POST)

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |
| out_trade_no | string | 是 | 商户订单号 |

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 1 | 成功 |
| -1 | 失败 |

**常见错误信息**:

- `Missing required parameters`: 缺少必需参数
- `Invalid signature`: 签名验证失败
- `Invalid merchant credentials`: 商户认证失败
- `Order not found`: 订单不存在
- `Order already paid`: 订单已支付
- `Invalid amount`: 金额格式错误
- `0 yuan purchase not allowed`: 不允许0元购

---

## 示例代码

### PHP 示例

```php
<?php
// 配置信息
$pid = '1001003549245339';
$key = 'your_merchant_key';
$api_url = 'http://your-domain.com';

// 订单信息
$params = [
    'pid' => $pid,
    'type' => 'alipay',
    'out_trade_no' => 'ORDER' . time(),
    'notify_url' => 'http://your-site.com/notify.php',
    'return_url' => 'http://your-site.com/return.php',
    'name' => '测试商品',
    'money' => '1.00',
    'sitename' => '我的网站'
];

// 生成签名
$sign_str = '';
ksort($params);
foreach ($params as $k => $v) {
    if ($v !== '' && $k !== 'sign' && $k !== 'sign_type') {
        $sign_str .= $k . '=' . $v . '&';
    }
}
$sign_str = rtrim($sign_str, '&') . $key;
$params['sign'] = md5($sign_str);
$params['sign_type'] = 'MD5';

// 发起请求
$response = file_get_contents($api_url . '/submit?' . http_build_query($params));
$result = json_decode($response, true);

if ($result['code'] == 1) {
    // 跳转到支付页面
    header('Location: ' . $result['payment_url']);
} else {
    echo '创建订单失败：' . $result['msg'];
}
?>
```

### Python 示例

```python
import hashlib
import requests
from urllib.parse import urlencode

# 配置信息
PID = '1001003549245339'
KEY = 'your_merchant_key'
API_URL = 'http://your-domain.com'

# 订单信息
params = {
    'pid': PID,
    'type': 'alipay',
    'out_trade_no': f'ORDER{int(time.time())}',
    'notify_url': 'http://your-site.com/notify',
    'return_url': 'http://your-site.com/return',
    'name': '测试商品',
    'money': '1.00',
    'sitename': '我的网站'
}

# 生成签名
sorted_params = sorted(params.items())
sign_str = '&'.join([f'{k}={v}' for k, v in sorted_params if v]) + KEY
params['sign'] = hashlib.md5(sign_str.encode()).hexdigest()
params['sign_type'] = 'MD5'

# 发起请求
response = requests.get(f'{API_URL}/submit', params=params)
result = response.json()

if result['code'] == 1:
    print(f"支付URL: {result['payment_url']}")
else:
    print(f"创建订单失败：{result['msg']}")
```

### JavaScript 示例

```javascript
// Node.js
const crypto = require('crypto');
const axios = require('axios');

// 配置信息
const PID = '1001003549245339';
const KEY = 'your_merchant_key';
const API_URL = 'http://your-domain.com';

// 订单信息
const params = {
    pid: PID,
    type: 'alipay',
    out_trade_no: `ORDER${Date.now()}`,
    notify_url: 'http://your-site.com/notify',
    return_url: 'http://your-site.com/return',
    name: '测试商品',
    money: '1.00',
    sitename: '我的网站'
};

// 生成签名
const sortedKeys = Object.keys(params).sort();
const signStr = sortedKeys
    .map(key => `${key}=${params[key]}`)
    .join('&') + KEY;
params.sign = crypto.createHash('md5').update(signStr).digest('hex');
params.sign_type = 'MD5';

// 发起请求
axios.get(`${API_URL}/submit`, { params })
    .then(response => {
        const result = response.data;
        if (result.code === 1) {
            console.log(`支付URL: ${result.payment_url}`);
        } else {
            console.log(`创建订单失败：${result.msg}`);
        }
    })
    .catch(error => console.error(error));
```

---

## 测试工具

### cURL 测试

```bash
# 查询商户信息
curl "http://localhost:8080/api?action=query&pid=YOUR_PID&key=YOUR_KEY"

# 创建订单
curl -X POST "http://localhost:8080/submit" \
  -d "pid=YOUR_PID" \
  -d "type=alipay" \
  -d "out_trade_no=TEST001" \
  -d "name=测试商品" \
  -d "money=1.00" \
  -d "notify_url=http://example.com/notify" \
  -d "return_url=http://example.com/return" \
  -d "sign=YOUR_SIGN"

# 查询订单
curl "http://localhost:8080/api/order?pid=YOUR_PID&out_trade_no=TEST001"
```

---

## 注意事项

1. **安全性**:
   - 请妥善保管商户密钥
   - 建议使用HTTPS协议
   - 验证异步通知的签名

2. **订单号**:
   - 商户订单号必须唯一
   - 建议使用时间戳+随机数

3. **金额**:
   - 金额必须大于0
   - 精确到分（小数点后两位）

4. **回调**:
   - 异步通知可能重复发送
   - 商户需做幂等性处理
   - 必须返回 `success` 表示接收成功

5. **超时**:
   - 订单默认5分钟超时
   - 超时订单会被自动清理

---

**更新时间**: 2024-01-15  
**版本**: v1.0.0

