# 易支付/码支付兼容性说明

## 签名算法

本系统完全兼容易支付（YiPay）和码支付（CodePay）的MD5签名标准。

### 签名生成流程

1. **过滤参数**：移除空值参数和 `sign`、`sign_type` 参数
2. **排序**：按参数名ASCII码升序排序
3. **拼接**：使用 `key1=value1&key2=value2` 格式拼接
4. **加密钥**：在字符串末尾拼接商户密钥
5. **MD5加密**：计算MD5并转为小写32位字符串

### 签名示例

```
参数：
  pid: 1001001276912812
  type: alipay
  out_trade_no: TEST123456
  notify_url: http://example.com/notify
  return_url: http://example.com/return
  name: 测试商品
  money: 0.01

商户密钥: f872e1c662d41cf218b5dfa8328ae455

签名字符串（排序后）:
money=0.01&name=测试商品&notify_url=http://example.com/notify&out_trade_no=TEST123456&pid=1001001276912812&return_url=http://example.com/return&type=alipay

加上商户密钥:
money=0.01&name=测试商品&notify_url=http://example.com/notify&out_trade_no=TEST123456&pid=1001001276912812&return_url=http://example.com/return&type=alipayf872e1c662d41cf218b5dfa8328ae455

MD5签名:
2fbd7fec465c508d33d815f420f02a3d
```

## API接口

### 1. 创建订单

**接口地址**：
- `/submit` （GET/POST）
- `/submit.php` （GET/POST，易支付兼容）

**请求参数**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| pid | string | 是 | 商户ID |
| type | string | 是 | 支付方式（alipay/wxpay） |
| out_trade_no | string | 是 | 商户订单号（唯一） |
| notify_url | string | 是 | 异步通知地址 |
| return_url | string | 是 | 同步跳转地址 |
| name | string | 是 | 商品名称 |
| money | string | 是 | 支付金额（元） |
| sitename | string | 否 | 网站名称 |
| sign | string | 是 | 签名 |
| sign_type | string | 否 | 签名类型（默认MD5） |

**返回**：
- 成功：显示支付页面（HTML）
- 失败：显示错误页面

### 2. 订单查询

**接口地址**：
- `/api/query` （GET/POST）
- `/api/query.php` （GET/POST，易支付兼容）

**请求参数**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| pid | string | 是 | 商户ID |
| trade_no | string | 是 | 订单号 |
| out_trade_no | string | 否 | 商户订单号 |
| sign | string | 是 | 签名 |

**返回JSON**：

```json
{
  "code": 1,
  "msg": "success",
  "trade_no": "20241023123456001",
  "out_trade_no": "TEST123456",
  "type": "alipay",
  "name": "测试商品",
  "money": "0.01",
  "trade_status": "TRADE_SUCCESS",
  "pay_time": "2024-10-23 12:00:00"
}
```

### 3. 异步通知

**回调地址**：商户在创建订单时指定的 `notify_url`

**通知参数**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| pid | string | 商户ID |
| trade_no | string | 系统订单号 |
| out_trade_no | string | 商户订单号 |
| type | string | 支付方式 |
| name | string | 商品名称 |
| money | string | 支付金额 |
| trade_status | string | 交易状态（TRADE_SUCCESS） |
| sign | string | 签名 |

**商户处理**：
1. 验证签名
2. 处理订单
3. 返回 `success` 或 `fail`

## 兼容性特性

### ✅ 已支持

- [x] MD5签名算法
- [x] 大小写不敏感签名比对
- [x] `.php` 后缀路由支持
- [x] URL双斜杠容错
- [x] GET/POST双方式支持
- [x] 异步通知机制
- [x] 订单查询接口
- [x] 订单状态回调

### 🔧 易支付完整兼容

| 功能 | 易支付 | 本系统 | 状态 |
|------|--------|--------|------|
| MD5签名 | ✓ | ✓ | ✅ 完全兼容 |
| 创建订单 | /submit | /submit | ✅ 完全兼容 |
| PHP后缀 | /submit.php | /submit.php | ✅ 完全兼容 |
| 异步通知 | ✓ | ✓ | ✅ 完全兼容 |
| 订单查询 | /api/query | /api/query | ✅ 完全兼容 |
| 返回格式 | JSON | JSON | ✅ 完全兼容 |

## 快速集成

### PHP示例

```php
<?php
// 商户信息
$pid = '1001001276912812';
$key = 'f872e1c662d41cf218b5dfa8328ae455';

// 订单信息
$params = array(
    'pid' => $pid,
    'type' => 'alipay',
    'out_trade_no' => 'ORDER' . time(),
    'notify_url' => 'http://your-domain.com/notify.php',
    'return_url' => 'http://your-domain.com/return.php',
    'name' => '测试商品',
    'money' => '0.01',
    'sign_type' => 'MD5'
);

// 生成签名
function generate_sign($params, $key) {
    // 过滤空值和sign参数
    $filtered = array_filter($params, function($v, $k) {
        return $v !== '' && $k !== 'sign' && $k !== 'sign_type';
    }, ARRAY_FILTER_USE_BOTH);
    
    // 排序
    ksort($filtered);
    
    // 拼接
    $sign_str = http_build_query($filtered);
    $sign_str = urldecode($sign_str);
    
    // 加密钥
    $sign_str .= $key;
    
    // MD5
    return strtolower(md5($sign_str));
}

$params['sign'] = generate_sign($params, $key);

// 构建支付URL
$payment_url = 'http://your-payment-gateway.com/submit?' . http_build_query($params);

// 跳转到支付页面
header('Location: ' . $payment_url);
?>
```

### Python示例

```python
import hashlib
import urllib.parse
import time

# 商户信息
PID = '1001001276912812'
KEY = 'f872e1c662d41cf218b5dfa8328ae455'

# 订单信息
params = {
    'pid': PID,
    'type': 'alipay',
    'out_trade_no': f'ORDER{int(time.time())}',
    'notify_url': 'http://your-domain.com/notify',
    'return_url': 'http://your-domain.com/return',
    'name': '测试商品',
    'money': '0.01',
    'sign_type': 'MD5'
}

# 生成签名
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
    
    # MD5
    return hashlib.md5(sign_str.encode()).hexdigest().lower()

params['sign'] = generate_sign(params, KEY)

# 构建支付URL
payment_url = f"http://your-payment-gateway.com/submit?{urllib.parse.urlencode(params)}"

print(payment_url)
```

## 测试工具

使用提供的测试脚本：

```bash
# 生成测试支付URL
python3 test_payment.py

# 查看服务器日志
tail -f logs/alimpay.log
```

## 注意事项

1. **签名验证**：所有接口请求必须包含正确的签名
2. **金额格式**：金额必须为正数，支持小数点后2位
3. **订单号唯一性**：同一商户的 `out_trade_no` 必须唯一
4. **回调处理**：异步通知可能会重复发送，请做好幂等性处理
5. **编码格式**：使用 UTF-8 编码

## 技术支持

如有问题，请查看日志文件：
- 应用日志：`logs/alimpay.log`
- 签名调试：启用 DEBUG 级别日志可查看详细签名验证信息

