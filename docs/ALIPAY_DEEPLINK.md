# 支付宝直接拉起支付功能

## 功能说明

支付宝直接拉起支付功能允许在移动端浏览器中直接打开支付宝应用，并自动填充二维码ID、金额和备注信息，从而简化支付流程。

## 配置方法

### 1. 配置二维码ID

在 `configs/config.yaml` 中添加支付宝二维码ID：

```yaml
payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
    qr_code_id: "fkx12345678901234"  # 您的支付宝二维码ID
    amount_offset: 0.01
    match_tolerance: 300
    payment_timeout: 300
```

### 2. 获取二维码ID

支付宝二维码ID可以从您的收款二维码URL中获取，格式通常为：
- `https://qr.alipay.com/fkx12345678901234`
- 其中 `fkx12345678901234` 就是二维码ID

## 使用方式

### 方式一：通过API接口生成深链接

#### 请求

```http
GET /alipay/link?amount=1.23&remark=测试订单
```

**参数说明：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| qr_code_id | string | 否 | 支付宝二维码ID，不填则使用配置中的默认值 |
| amount | float | 否 | 支付金额，单位：元 |
| remark | string | 否 | 备注信息 |

#### 响应

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "qr_code_id": "fkx12345678901234",
  "amount": 1.23,
  "remark": "测试订单",
  "alipay_deep_link": "alipays://platformapi/startapp?appId=20000056&url=https%3A%2F%2Fqr.alipay.com%2Ffkx12345678901234%3Famount%3D1.23%26remark%3D%E6%B5%8B%E8%AF%95%E8%AE%A2%E5%8D%95",
  "usage": "在移动端浏览器中访问此链接可直接拉起支付宝进行支付"
}
```

### 方式二：直接重定向到支付宝

#### 请求

```http
GET /alipay/pay?amount=1.23&remark=测试订单
```

该接口会直接重定向到支付宝深链接，适合在移动端浏览器中直接使用。

### 方式三：通过支付订单自动生成

当创建支付订单时，如果配置了 `qr_code_id`，系统会自动在响应中包含 `alipay_deep_link` 字段：

```http
POST /submit
```

响应中会包含：

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "trade_no": "20240115120000123456",
  "payment_url": "http://your-domain.com/pay?trade_no=xxx&amount=1.23",
  "qr_code": "data:image/png;base64,...",
  "alipay_deep_link": "alipays://platformapi/startapp?appId=20000056&url=..."
}
```

### 方式四：在支付页面使用

访问支付页面时，如果配置了 `qr_code_id`，页面会包含 `alipay_deep_link` 数据，前端可以显示"直接打开支付宝"按钮。

## 深链接格式说明

生成的支付宝深链接格式为：

```
alipays://platformapi/startapp?appId=20000056&url=https%3A%2F%2Fqr.alipay.com%2F{qrCodeId}%3Famount%3D{amount}%26remark%3D{remark}
```

**参数说明：**

- `appId=20000056`: 支付宝转账应用ID（固定值）
- `url`: URL编码后的支付宝二维码链接
  - `qrCodeId`: 您的支付宝二维码ID
  - `amount`: 支付金额（可选）
  - `remark`: 备注信息（可选）

## 前端集成示例

### HTML + JavaScript

```html
<!DOCTYPE html>
<html>
<head>
    <title>支付宝支付</title>
</head>
<body>
    <button onclick="openAlipay()">打开支付宝支付</button>
    
    <script>
    function openAlipay() {
        const amount = 1.23;
        const remark = '测试订单';
        
        // 获取深链接
        fetch(`/alipay/link?amount=${amount}&remark=${encodeURIComponent(remark)}`)
            .then(response => response.json())
            .then(data => {
                if (data.code === 1) {
                    // 直接跳转到支付宝
                    window.location.href = data.alipay_deep_link;
                } else {
                    alert('生成支付链接失败：' + data.msg);
                }
            });
    }
    </script>
</body>
</html>
```

### 移动端H5页面

```html
<a href="alipays://platformapi/startapp?appId=20000056&url=https%3A%2F%2Fqr.alipay.com%2Ffkx12345678901234%3Famount%3D1.23">
    打开支付宝支付
</a>
```

## 注意事项

1. **平台限制**：深链接仅在移动端（iOS/Android）的支付宝应用中有效
2. **浏览器支持**：某些浏览器可能会拦截深链接跳转，需要用户手动确认
3. **安全性**：建议在生成深链接时进行金额和参数验证
4. **二维码ID**：确保配置的二维码ID是有效的支付宝收款码ID
5. **金额限制**：
   - 最小金额：0.01元
   - 最大金额：99999.99元
   - 金额必须大于0（防止0元购）

## 测试方法

### 1. 使用curl测试API

```bash
# 生成深链接
curl "http://localhost:8080/alipay/link?amount=0.01&remark=test"

# 直接重定向（会返回302）
curl -L "http://localhost:8080/alipay/pay?amount=0.01&remark=test"
```

### 2. 在移动设备上测试

1. 确保手机安装了支付宝应用
2. 在移动浏览器中访问生成的深链接
3. 系统会自动拉起支付宝应用并填充支付信息

### 3. 使用二维码测试

可以将深链接转换为二维码，用户扫码后自动打开支付宝：

```bash
# 生成包含深链接的二维码
qrencode -o alipay_link.png "alipays://platformapi/startapp?appId=20000056&url=https%3A%2F%2Fqr.alipay.com%2Ffkx12345678901234%3Famount%3D1.23"
```

## 故障排查

### 问题1：深链接无法打开支付宝

**可能原因：**
- 未安装支付宝应用
- 二维码ID配置错误
- URL编码问题

**解决方法：**
- 检查 `qr_code_id` 配置是否正确
- 确认移动设备已安装支付宝
- 检查生成的URL是否正确编码

### 问题2：金额没有自动填充

**可能原因：**
- 金额参数格式错误
- 二维码不支持金额参数

**解决方法：**
- 确认金额格式为数字且大于0
- 验证支付宝二维码是否支持金额参数

## 相关链接

- [支付宝开放平台](https://open.alipay.com/)
- [支付宝H5支付文档](https://opendocs.alipay.com/open/203/105285)
- [AliMPay项目主页](https://github.com/chanhanzhan/AliMPay)
