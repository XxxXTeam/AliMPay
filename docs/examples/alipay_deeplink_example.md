# 支付宝深链接功能使用示例

## 示例1：基础配置

在 `configs/config.yaml` 中配置：

```yaml
payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
    qr_code_id: "fkx12345678901234"  # 替换为您的支付宝二维码ID
```

## 示例2：通过API生成深链接

### 请求示例

```bash
# 使用默认配置的二维码ID
curl "http://localhost:8080/alipay/link?amount=1.23&remark=测试订单"

# 使用自定义二维码ID
curl "http://localhost:8080/alipay/link?qr_code_id=custom123&amount=5.00&remark=自定义订单"

# 只生成二维码链接，不带金额
curl "http://localhost:8080/alipay/link"
```

### 响应示例

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

## 示例3：移动端H5页面集成

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>支付宝支付</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
        }
        .pay-button {
            background-color: #1677FF;
            color: white;
            padding: 15px 30px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            width: 100%;
            margin-top: 20px;
        }
    </style>
</head>
<body>
    <h1>订单支付</h1>
    
    <div class="order-info">
        <p><strong>订单号：</strong>TEST20240115001</p>
        <p><strong>商品名称：</strong>测试商品</p>
        <p><strong>支付金额：</strong>¥1.23</p>
    </div>
    
    <button class="pay-button" onclick="openAlipayPay()">
        打开支付宝支付
    </button>
    
    <script>
    function openAlipayPay() {
        const amount = 1.23;
        const remark = '测试订单';
        
        // 直接重定向到支付宝
        window.location.href = `/alipay/pay?amount=${amount}&remark=${encodeURIComponent(remark)}`;
    }
    </script>
</body>
</html>
```

## 示例4：Go语言客户端调用

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
)

type AlipayLinkResponse struct {
    Code          int     `json:"code"`
    Msg           string  `json:"msg"`
    AlipayDeepLink string `json:"alipay_deep_link"`
}

func generateAlipayLink(baseURL string, amount float64, remark string) (string, error) {
    reqURL := fmt.Sprintf("%s/alipay/link?amount=%.2f&remark=%s",
        baseURL, amount, url.QueryEscape(remark))

    resp, err := http.Get(reqURL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    var result AlipayLinkResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return "", err
    }

    return result.AlipayDeepLink, nil
}
```

## 注意事项

1. **移动端专用**：深链接仅在移动设备上有效
2. **支付宝应用**：确保用户设备安装了支付宝应用
3. **HTTPS要求**：生产环境建议使用HTTPS协议
