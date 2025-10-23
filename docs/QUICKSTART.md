# AliMPay 快速开始 / Quick Start

本指南将帮助您在 10 分钟内快速部署并运行 AliMPay 支付系统。

This guide will help you deploy and run AliMPay payment system in 10 minutes.

---

## 📋 准备工作 / Prerequisites

在开始之前，请确保您已准备好：

Before starting, ensure you have:

- ✅ 支付宝开放平台账号 / Alipay Open Platform account
- ✅ 支付宝商家收款码 / Alipay merchant collection QR code
- ✅ 服务器或本地开发环境 / Server or local development environment

---

## 🚀 三步快速部署 / Three-Step Quick Deployment

### 步骤 1: 获取支付宝配置 / Step 1: Get Alipay Configuration

#### 1.1 登录支付宝开放平台 / Login to Alipay Open Platform

访问 https://open.alipay.com 并登录

#### 1.2 创建应用 / Create Application

1. 进入"控制台" > "我的应用"
2. 点击"创建应用"
3. 选择"网页/移动应用"
4. 填写应用信息并提交审核

#### 1.3 获取 AppID

审核通过后，在应用详情页面可以看到 **AppID**（例如：2021001234567890）

#### 1.4 生成密钥对 / Generate Key Pair

**下载密钥生成工具：**
- Windows: https://ideservice.alipay.com/ide/getPluginUrl.htm?clientType=assistant&platform=win
- macOS: https://ideservice.alipay.com/ide/getPluginUrl.htm?clientType=assistant&platform=mac

**生成步骤：**
1. 打开支付宝开放平台开发助手
2. 选择"RSA2(SHA256)密钥"
3. 点击"生成密钥"
4. 保存应用私钥（应用私钥.txt）
5. 复制应用公钥

#### 1.5 上传应用公钥 / Upload Application Public Key

1. 在应用详情页面找到"开发信息"
2. 点击"设置"上传应用公钥
3. 复制并保存支付宝公钥

#### 1.6 获取用户ID / Get User ID

1. 点击开放平台右上角头像
2. 进入"账号中心"
3. 查看并复制"账号UID"（例如：2088123456789012）

#### 1.7 下载经营码 / Download Business QR Code

1. 登录支付宝商家中心：https://b.alipay.com
2. 进入"店铺管理" > "收款码"
3. 下载"商家经营收款码"保存为图片

---

### 步骤 2: 部署 AliMPay / Step 2: Deploy AliMPay

选择一种部署方式：

Choose one deployment method:

#### 方式 A: 使用 Docker（推荐新手）/ Using Docker (Recommended for Beginners)

```bash
# 1. 克隆代码
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay

# 2. 准备配置文件
cp configs/config.example.yaml configs/config.yaml

# 3. 编辑配置文件（使用您喜欢的编辑器）
vim configs/config.yaml
# 或
nano configs/config.yaml
```

**填写关键配置：**
```yaml
alipay:
  app_id: "2021001234567890"              # 您的 AppID
  private_key: "MIIEvQIBA..."             # 您的应用私钥
  alipay_public_key: "MIIBIjANBg..."      # 支付宝公钥
  transfer_user_id: "2088123456789012"   # 您的用户ID

payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
```

```bash
# 4. 复制经营码图片
cp /path/to/your/qr_code.png qrcode/business_qr.png

# 5. 使用 Docker Compose 启动
docker-compose up -d

# 6. 查看日志
docker-compose logs -f
```

#### 方式 B: 直接运行（适合开发测试）/ Direct Run (For Development/Testing)

```bash
# 1. 确保已安装 Go 1.23+
go version

# 2. 克隆代码
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay

# 3. 下载依赖
go mod download

# 4. 准备配置文件
cp configs/config.example.yaml configs/config.yaml
vim configs/config.yaml  # 按上述方式填写配置

# 5. 复制经营码图片
cp /path/to/your/qr_code.png qrcode/business_qr.png

# 6. 编译运行
make build
./alimpay -config=./configs/config.yaml
```

---

### 步骤 3: 验证部署 / Step 3: Verify Deployment

#### 3.1 检查服务状态 / Check Service Status

```bash
# 健康检查
curl http://localhost:8080/health

# 预期输出：
# {"status":"ok","timestamp":"2024-01-15T12:00:00Z"}
```

#### 3.2 获取商户信息 / Get Merchant Info

服务启动后，在日志中查找商户ID和密钥：

After service starts, find merchant ID and key in logs:

```bash
# Docker
docker-compose logs | grep "Merchant"

# 直接运行
tail -f logs/alimpay.log | grep "Merchant"
```

输出示例：
```
Merchant ID: 1001003549245339
Merchant Key: f872e1c662d41cf218b5dfa8328ae455
```

**保存这两个值！您将在API调用时使用。**

#### 3.3 访问管理后台 / Access Admin Dashboard

在浏览器中访问：
```
http://localhost:8080/admin/dashboard
```

您应该能看到订单管理界面。

---

## 🧪 测试支付 / Test Payment

### 使用测试脚本 / Using Test Script

项目包含了测试脚本，可以快速生成测试订单：

```bash
# 编辑测试脚本，填入您的商户信息
vim test_payment.py

# 修改以下变量：
# PID = "您的商户ID"
# KEY = "您的商户密钥"
# API_URL = "http://localhost:8080"

# 运行测试
python3 test_payment.py
```

脚本会输出支付链接，访问该链接即可看到支付页面。

### 手动创建订单 / Manual Order Creation

#### 使用 cURL 测试

```bash
# 1. 准备参数
PID="您的商户ID"
KEY="您的商户密钥"
OUT_TRADE_NO="TEST$(date +%s)"

# 2. 生成签名（使用 Python）
python3 << EOF
import hashlib
params = {
    'money': '0.01',
    'name': '测试商品',
    'notify_url': 'http://example.com/notify',
    'out_trade_no': '${OUT_TRADE_NO}',
    'pid': '${PID}',
    'return_url': 'http://example.com/return',
    'type': 'alipay'
}
sign_str = '&'.join([f'{k}={params[k]}' for k in sorted(params.keys())]) + '${KEY}'
print(hashlib.md5(sign_str.encode()).hexdigest())
EOF

# 3. 使用生成的签名创建订单
SIGN="生成的签名"
curl -X POST "http://localhost:8080/submit" \
  -d "pid=${PID}" \
  -d "type=alipay" \
  -d "out_trade_no=${OUT_TRADE_NO}" \
  -d "name=测试商品" \
  -d "money=0.01" \
  -d "notify_url=http://example.com/notify" \
  -d "return_url=http://example.com/return" \
  -d "sign=${SIGN}"
```

---

## 💡 接下来做什么？/ What's Next?

### 1. 集成到您的应用 / Integrate into Your Application

参考 [接入教程](INTEGRATION.md) 了解如何在您的应用中调用 AliMPay API。

### 2. 配置生产环境 / Configure Production Environment

参考 [部署教程](DEPLOYMENT.md) 了解：
- 使用 Systemd 管理服务
- 配置 Nginx 反向代理
- 启用 HTTPS
- 性能优化

### 3. 自定义配置 / Customize Configuration

查看 [配置文件](../configs/config.example.yaml) 了解所有可用配置项。

### 4. 了解 API / Learn API

阅读 [API 文档](API.md) 了解所有可用接口。

---

## 🔍 常见问题 / Common Issues

### 问题 1: 服务启动失败

**解决方法：**
1. 检查配置文件格式是否正确（YAML 语法）
2. 检查必填字段是否都已填写
3. 查看日志获取详细错误信息

### 问题 2: 无法访问服务

**解决方法：**
1. 检查防火墙是否开放 8080 端口
   ```bash
   # Ubuntu/Debian
   sudo ufw allow 8080
   
   # CentOS/RHEL
   sudo firewall-cmd --add-port=8080/tcp --permanent
   sudo firewall-cmd --reload
   ```

2. 如果使用 Docker，检查端口映射是否正确

### 问题 3: 支付后没有跳转

**解决方法：**
1. 确保监控服务已启用（配置文件中 `monitor.enabled: true`）
2. 检查支付宝 API 权限是否已开通
3. 查看日志中是否有错误信息

更多问题请参考 [常见问题文档](FAQ.md)。

---

## 📚 延伸阅读 / Further Reading

- [完整部署教程](DEPLOYMENT.md)
- [接入教程](INTEGRATION.md)
- [API 文档](API.md)
- [常见问题](FAQ.md)
- [贡献指南](../CONTRIBUTING.md)

---

## 🆘 获取帮助 / Get Help

如果遇到问题：

1. 查看 [常见问题文档](FAQ.md)
2. 搜索 [GitHub Issues](https://github.com/chanhanzhan/AliMPay/issues)
3. 提交新的 Issue
4. 发送邮件至 support@openel.top

---

## ⚡ 快速命令参考 / Quick Command Reference

```bash
# 查看服务状态
curl http://localhost:8080/health

# 查看日志（Docker）
docker-compose logs -f

# 查看日志（直接运行）
tail -f logs/alimpay.log

# 重启服务（Docker）
docker-compose restart

# 重启服务（Systemd）
sudo systemctl restart alimpay

# 停止服务（Docker）
docker-compose down

# 停止服务（直接运行）
pkill alimpay

# 备份数据库
cp data/alimpay.db data/alimpay.db.backup.$(date +%Y%m%d)

# 清理过期订单
curl "http://localhost:8080/health?action=cleanup"

# 手动触发监控
curl "http://localhost:8080/health?action=monitor"
```

---

**祝您使用愉快！/ Happy Using!** 🎉

如果觉得项目有帮助，欢迎给个 ⭐️ Star！

If you find the project helpful, feel free to give it a ⭐️ Star!
