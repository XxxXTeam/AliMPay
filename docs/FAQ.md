# AliMPay 常见问题 / FAQ

本文档收录了使用 AliMPay 过程中的常见问题和解决方案。

This document collects common questions and solutions when using AliMPay.

---

## 目录 / Table of Contents

- [安装与部署](#安装与部署--installation-and-deployment)
- [配置相关](#配置相关--configuration)
- [支付功能](#支付功能--payment-features)
- [API接口](#api接口--api-interface)
- [签名验证](#签名验证--signature-verification)
- [错误排查](#错误排查--troubleshooting)
- [性能优化](#性能优化--performance-optimization)

---

## 安装与部署 / Installation and Deployment

### Q1: 支持哪些操作系统？

**A:** AliMPay 支持以下操作系统：
- Linux (Ubuntu 20.04+, CentOS 7+, Debian 10+)
- macOS (10.15+)
- Windows (10/11/Server 2019+)
- 支持 Docker 容器部署，可在任何支持 Docker 的平台运行

### Q2: 需要什么样的服务器配置？

**A:** 
**最低配置：**
- CPU: 1核
- 内存: 512MB
- 硬盘: 1GB

**推荐配置：**
- CPU: 2核
- 内存: 2GB
- 硬盘: 10GB SSD
- 带宽: 1Mbps+

### Q3: 如何编译项目？

**A:** 
```bash
# 确保安装了 Go 1.23+
go version

# 克隆代码
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay

# 下载依赖
go mod download

# 编译
make build
# 或者
go build -o alimpay ./cmd/alimpay
```

### Q4: Docker 部署时如何持久化数据？

**A:** 使用 volume 挂载数据目录：
```bash
docker run -d \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/qrcode:/app/qrcode \
  alimpay:latest
```

### Q5: 如何升级到新版本？

**A:** 
**源码部署：**
```bash
# 备份数据
cp -r data data.backup

# 拉取新代码
git pull

# 重新编译
make build

# 重启服务
systemctl restart alimpay
```

**Docker部署：**
```bash
# 拉取新镜像
docker pull chanhanzhan/alimpay:latest

# 停止旧容器
docker stop alimpay
docker rm alimpay

# 启动新容器（数据已持久化，不会丢失）
docker run -d --name alimpay ... alimpay:latest
```

---

## 配置相关 / Configuration

### Q6: 如何获取支付宝 AppID 和密钥？

**A:** 
1. 登录 [支付宝开放平台](https://open.alipay.com)
2. 进入"控制台" > "我的应用"
3. 创建网页/移动应用
4. 获取 AppID
5. 使用密钥生成工具生成应用公钥和私钥
6. 上传应用公钥到开放平台
7. 获取支付宝公钥

详细教程：https://opendocs.alipay.com/common/02kipk

### Q7: 如何获取支付宝用户ID (transfer_user_id)？

**A:** 
**方式一：** 通过开放平台查看
1. 登录支付宝开放平台
2. 点击右上角头像
3. 进入"账号中心"
4. 查看"账号UID"

**方式二：** 通过API查询
```bash
# 需要先配置好其他参数
./alimpay --get-user-id
```

### Q8: 经营码图片从哪里获取？

**A:** 
1. 登录支付宝商家中心：https://b.alipay.com
2. 进入"店铺管理" > "收款码"
3. 下载"商家经营收款码"
4. 将图片保存为 `qrcode/business_qr.png`

### Q9: 如何获取收款码ID (qr_code_id)？

**A:** 
1. 用手机支付宝扫描经营码
2. 查看浏览器地址栏
3. 找到类似 `https://qr.alipay.com/fkx123456` 的链接
4. `fkx123456` 就是收款码ID
5. 填入配置文件的 `qr_code_id` 字段

**注意：** 如果不填写此字段，支付页面将不显示"拉起支付宝"按钮，但不影响扫码支付。

### Q10: 配置文件修改后需要重启吗？

**A:** 是的，配置文件修改后必须重启服务才能生效：
```bash
# Systemd
systemctl restart alimpay

# Docker
docker restart alimpay

# 手动运行
# Ctrl+C 停止后重新运行
./alimpay -config=./configs/config.yaml
```

---

## 支付功能 / Payment Features

### Q11: 经营码模式和转账模式有什么区别？

**A:** 

| 特性 | 经营码模式 | 转账模式 |
|------|-----------|---------|
| 二维码 | 固定二维码 | 每单独立二维码 |
| 到账速度 | 即时到账 | 即时到账 |
| 匹配方式 | 金额+时间 | 订单号 |
| 手续费 | 无（直接到账） | 无（转账免费） |
| 推荐度 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

**推荐使用经营码模式**，配置简单，到账快，用户体验好。

### Q12: 为什么实际支付金额和订单金额不一样？

**A:** 这是经营码模式的特性。为了避免同一时间多个订单金额相同导致匹配错误，系统会在订单金额基础上加一个偏移量（默认0.01元）。

例如：
- 订单金额：1.00 元
- 实际支付：1.01 元

可以在配置文件中调整偏移量：
```yaml
payment:
  business_qr_mode:
    amount_offset: 0.01  # 调整此值
```

### Q13: 支付成功后没有跳转怎么办？

**A:** 
1. **检查监控服务是否启用**
   ```yaml
   monitor:
     enabled: true  # 确保为 true
   ```

2. **查看日志**
   ```bash
   tail -f logs/alimpay.log | grep "monitor"
   ```

3. **检查支付宝API权限**
   - 确保已开通"查询对账单下载地址"权限
   - 确保已开通"单笔转账"权限

4. **手动触发监控**
   ```bash
   curl "http://localhost:8080/health?action=monitor"
   ```

### Q14: 订单超时时间可以修改吗？

**A:** 可以，在配置文件中修改：
```yaml
payment:
  order_timeout: 300  # 单位：秒，默认5分钟
  payment_timeout: 300  # 支付超时时间
```

### Q15: 如何查看未支付的订单？

**A:** 
**方式一：** 访问管理后台
```
http://your-domain.com/admin/dashboard
```

**方式二：** 通过API查询
```bash
curl "http://localhost:8080/api/orders?pid=YOUR_PID&key=YOUR_KEY"
```

---

## API接口 / API Interface

### Q16: API 支持哪些请求方式？

**A:** 大部分 API 同时支持 GET 和 POST 请求：
```bash
# GET 方式
curl "http://localhost:8080/submit?pid=xxx&..."

# POST 方式
curl -X POST "http://localhost:8080/submit" -d "pid=xxx&..."
```

### Q17: 如何测试 API 接口？

**A:** 
**使用 curl：**
```bash
# 测试创建订单（需要替换参数和签名）
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

**使用测试脚本：**
```bash
# Python 测试脚本
python3 test_payment.py
```

### Q18: 异步回调通知会重试吗？

**A:** 会的。如果回调地址没有返回 `success` 或 `ok`，系统会重试通知（目前实现为单次通知，计划增加重试机制）。

**商户端建议：**
- 实现幂等性处理（防止重复处理）
- 收到通知后立即返回 `success`
- 在异步处理业务逻辑

### Q19: 如何关闭订单？

**A:** 
```bash
curl -X POST "http://localhost:8080/api/close" \
  -d "pid=YOUR_PID" \
  -d "key=YOUR_KEY" \
  -d "out_trade_no=ORDER123"
```

### Q20: 返回的二维码是什么格式？

**A:** 返回的 `qr_code` 字段是 Base64 编码的 PNG 图片：
```json
{
  "qr_code": "data:image/png;base64,iVBORw0KGgo..."
}
```

可以直接在 HTML 中使用：
```html
<img src="data:image/png;base64,iVBORw0KGgo..." alt="支付二维码">
```

---

## 签名验证 / Signature Verification

### Q21: 签名验证总是失败？

**A:** 请逐步检查：

1. **参数排序是否正确**
   - 必须按照参数名的 ASCII 码升序排列
   - 例如：money < name < notify_url

2. **是否过滤了 sign 和 sign_type**
   ```python
   # 正确做法
   filtered = {k: v for k, v in params.items() 
               if v and k not in ['sign', 'sign_type']}
   ```

3. **URL 编码处理**
   - 签名时不应该进行 URL 编码
   - 直接拼接原始参数值

4. **商户密钥是否正确**
   - 从配置文件或日志中确认密钥
   - 注意不要有多余的空格

5. **MD5 是否转小写**
   ```python
   sign = hashlib.md5(sign_str.encode()).hexdigest().lower()
   ```

6. **调试技巧**
   ```python
   # 输出待签名字符串，与服务端对比
   print(f"待签名字符串: {sign_str}")
   print(f"计算的签名: {sign}")
   ```

### Q22: 有签名验证的在线工具吗？

**A:** 可以启用调试模式查看详细签名信息：
```yaml
logging:
  level: "debug"  # 改为 debug
```

重启服务后，日志会输出详细的签名验证过程。

### Q23: 不同语言的签名结果不一致？

**A:** 常见原因：
1. **字符编码不一致** - 统一使用 UTF-8
2. **URL 编码处理不同** - 签名时不要 URL 编码
3. **参数值有特殊字符** - 确保原样传递
4. **MD5 实现差异** - 确保转小写

**验证方法：**
使用相同参数在不同语言中生成签名，应该得到相同结果。

---

## 错误排查 / Troubleshooting

### Q24: 服务无法启动，报错 "address already in use"？

**A:** 端口被占用，解决方法：

**方式一：** 修改端口
```yaml
server:
  port: 8081  # 改为其他端口
```

**方式二：** 查找并关闭占用进程
```bash
# Linux/macOS
lsof -i :8080
kill <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Q25: 日志显示 "failed to query bills"？

**A:** 可能原因：
1. **支付宝API权限未开通**
   - 登录开放平台，检查是否开通"查询对账单下载地址"权限

2. **AppID 或密钥配置错误**
   - 检查配置文件中的支付宝配置

3. **网络连接问题**
   - 检查服务器是否可以访问支付宝API
   ```bash
   curl https://openapi.alipay.com
   ```

4. **时间不同步**
   - 检查服务器时间是否准确
   ```bash
   date
   # 同步时间
   ntpdate pool.ntp.org
   ```

### Q26: 数据库文件损坏如何恢复？

**A:** 
1. **从备份恢复**
   ```bash
   cp data/alimpay.db.backup data/alimpay.db
   ```

2. **SQLite 修复**
   ```bash
   sqlite3 data/alimpay.db ".recover" | sqlite3 data/alimpay_recovered.db
   mv data/alimpay_recovered.db data/alimpay.db
   ```

3. **重建数据库**
   ```bash
   rm data/alimpay.db
   make init
   ```

### Q27: Nginx 反向代理后无法访问？

**A:** 检查 Nginx 配置：

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

确保：
1. Nginx 可以访问后端服务（127.0.0.1:8080）
2. 防火墙允许 Nginx 访问
3. SELinux 配置正确（CentOS/RHEL）

### Q28: Docker 容器内无法访问外部网络？

**A:** 
1. **检查 Docker 网络配置**
   ```bash
   docker network ls
   docker network inspect bridge
   ```

2. **测试网络连通性**
   ```bash
   docker exec alimpay ping -c 3 openapi.alipay.com
   ```

3. **检查 DNS 配置**
   ```bash
   docker exec alimpay cat /etc/resolv.conf
   ```

---

## 性能优化 / Performance Optimization

### Q29: 如何提升系统性能？

**A:** 
1. **优化数据库连接池**
   ```yaml
   database:
     max_idle_conns: 20
     max_open_conns: 200
   ```

2. **调整监控间隔**
   ```yaml
   monitor:
     interval: 5  # 根据实际需求调整，不要过小
   ```

3. **使用生产模式**
   ```yaml
   server:
     mode: "release"
   logging:
     level: "warn"  # 或 "error"
   ```

4. **启用日志压缩**
   ```yaml
   logging:
     compress: true
   ```

5. **使用 SSD 存储**

6. **增加服务器配置**
   - 2核CPU + 2GB内存可支持较大并发

### Q30: 如何监控系统性能？

**A:** 
1. **查看健康状态**
   ```bash
   curl http://localhost:8080/health?action=status
   ```

2. **查看资源使用**
   ```bash
   # 进程资源
   top -p $(pgrep alimpay)
   
   # 内存使用
   ps aux | grep alimpay
   
   # 磁盘使用
   df -h
   du -sh data/ logs/
   ```

3. **日志分析**
   ```bash
   # 错误统计
   grep "ERROR" logs/alimpay.log | wc -l
   
   # 响应时间
   grep "latency" logs/alimpay.log | tail -20
   ```

### Q31: 大量订单时如何优化？

**A:** 
1. **定期清理过期订单**
   - 已自动启用，也可手动触发：
   ```bash
   curl "http://localhost:8080/health?action=cleanup"
   ```

2. **考虑使用 MySQL/PostgreSQL**
   - SQLite 适合中小规模，大规模建议使用 MySQL

3. **数据归档**
   - 定期将历史订单导出并删除

4. **使用 Redis 缓存**（未来版本支持）

---

## 安全相关 / Security

### Q32: 如何保护商户密钥安全？

**A:** 
1. **文件权限**
   ```bash
   chmod 600 configs/config.yaml
   ```

2. **不要提交到版本控制**
   ```bash
   echo "configs/config.yaml" >> .gitignore
   ```

3. **定期更换密钥**

4. **使用环境变量**（未来版本支持）

### Q33: 如何启用 HTTPS？

**A:** 推荐使用 Nginx 反向代理 + Let's Encrypt 证书，详见 [部署教程](DEPLOYMENT.md#https配置--https-configuration)。

### Q34: 如何限制 API 访问频率？

**A:** 建议在 Nginx 层面配置限流：
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
    ...
}
```

---

## 其他问题 / Other Questions

### Q35: 支持哪些支付方式？

**A:** 当前版本仅支持支付宝支付。未来计划支持：
- 微信支付
- 云闪付
- 其他支付方式

### Q36: 可以修改支付页面样式吗？

**A:** 可以，修改 `web/templates/` 目录下的 HTML 模板文件和 `web/static/` 目录下的 CSS 文件。

### Q37: 数据会丢失吗？

**A:** 
- SQLite 数据库文件保存在 `data/` 目录
- 建议定期备份数据库文件
- Docker 部署时使用 volume 挂载确保数据持久化

### Q38: 如何贡献代码？

**A:** 欢迎贡献！请参考 [贡献指南](../CONTRIBUTING.md)。

### Q39: 商业使用需要授权吗？

**A:** 本项目采用 MIT 协议开源，可免费用于商业用途，无需额外授权。

### Q40: 在哪里获取技术支持？

**A:** 
- **GitHub Issues**: https://github.com/chanhanzhan/AliMPay/issues
- **文档**: https://github.com/chanhanzhan/AliMPay/tree/main/docs
- **Email**: support@openel.top

---

**如果您的问题未在此列出，欢迎提交 Issue！**

**If your question is not listed here, feel free to submit an Issue!**
