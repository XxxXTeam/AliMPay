# AliMPay 部署教程 / Deployment Guide

本文档提供详细的部署指南，包含多种部署方式和生产环境最佳实践。

This document provides detailed deployment guide, including multiple deployment methods and production best practices.

---

## 目录 / Table of Contents

- [环境准备](#环境准备--environment-preparation)
- [本地部署](#本地部署--local-deployment)
- [Docker部署](#docker部署--docker-deployment)
- [Docker Compose部署](#docker-compose部署--docker-compose-deployment)
- [生产环境部署](#生产环境部署--production-deployment)
- [Nginx反向代理配置](#nginx反向代理配置--nginx-reverse-proxy)
- [HTTPS配置](#https配置--https-configuration)
- [监控与维护](#监控与维护--monitoring-and-maintenance)
- [常见问题](#常见问题--troubleshooting)

---

## 环境准备 / Environment Preparation

### 系统要求 / System Requirements

**最低配置 / Minimum:**
- CPU: 1核 / 1 Core
- 内存 / RAM: 512MB
- 硬盘 / Disk: 1GB
- 系统 / OS: Linux (Ubuntu 20.04+, CentOS 7+, Debian 10+) / macOS / Windows

**推荐配置 / Recommended:**
- CPU: 2核 / 2 Cores
- 内存 / RAM: 2GB
- 硬盘 / Disk: 10GB SSD
- 系统 / OS: Linux (Ubuntu 22.04 LTS)

### 软件依赖 / Software Dependencies

**必需 / Required:**
- Go 1.23 或更高版本 / Go 1.23 or higher (仅源码部署需要 / only for source deployment)
- Git

**可选 / Optional:**
- Docker 20.10+ (用于容器部署 / for container deployment)
- Docker Compose 2.0+ (用于编排部署 / for orchestrated deployment)
- Nginx (用于反向代理 / for reverse proxy)

---

## 本地部署 / Local Deployment

### 方式一：使用预编译二进制文件 / Method 1: Using Pre-compiled Binary

**步骤 / Steps:**

#### 1. 下载最新版本 / Download Latest Release

访问 [Releases 页面](https://github.com/chanhanzhan/AliMPay/releases) 下载适合你系统的版本：

Visit [Releases page](https://github.com/chanhanzhan/AliMPay/releases) to download the version for your system:

```bash
# Linux AMD64
wget https://github.com/chanhanzhan/AliMPay/releases/download/vX.X.X/alimpay-linux-amd64.tar.gz
tar -xzf alimpay-linux-amd64.tar.gz
cd alimpay-linux-amd64

# macOS
wget https://github.com/chanhanzhan/AliMPay/releases/download/vX.X.X/alimpay-darwin-amd64.tar.gz
tar -xzf alimpay-darwin-amd64.tar.gz
cd alimpay-darwin-amd64

# Windows
# 下载 alimpay-windows-amd64.zip 并解压
# Download alimpay-windows-amd64.zip and extract
```

#### 2. 配置文件 / Configuration

```bash
# 复制配置文件模板
# Copy configuration template
cp configs/config.example.yaml configs/config.yaml

# 编辑配置文件，填写支付宝相关信息
# Edit configuration file, fill in Alipay information
vim configs/config.yaml  # 或使用其他编辑器 / or use other editors
```

**必需配置项 / Required Configuration:**

```yaml
alipay:
  app_id: "你的支付宝应用ID / Your Alipay App ID"
  private_key: "你的应用私钥 / Your Application Private Key"
  alipay_public_key: "支付宝公钥 / Alipay Public Key"
  transfer_user_id: "收款支付宝用户ID / Recipient Alipay User ID"

payment:
  business_qr_mode:
    enabled: true
    qr_code_path: "./qrcode/business_qr.png"
```

#### 3. 准备经营码（如果使用经营码模式）/ Prepare Business QR Code

```bash
# 将您的支付宝经营码图片保存到指定位置
# Save your Alipay business QR code image to specified location
cp /path/to/your/business_qr.png qrcode/business_qr.png
```

#### 4. 启动服务 / Start Service

```bash
# 赋予执行权限 (Linux/macOS)
# Grant execute permission (Linux/macOS)
chmod +x alimpay

# 启动服务
# Start service
./alimpay -config=./configs/config.yaml

# 后台运行 (推荐使用 systemd 或 supervisor)
# Run in background (recommend using systemd or supervisor)
nohup ./alimpay -config=./configs/config.yaml > logs/alimpay.log 2>&1 &
```

#### 5. 验证部署 / Verify Deployment

```bash
# 检查服务状态
# Check service status
curl http://localhost:8080/health

# 预期输出 / Expected output:
# {"status":"ok","timestamp":"2024-01-15T12:00:00Z"}
```

---

### 方式二：从源码编译 / Method 2: Build from Source

#### 1. 克隆代码仓库 / Clone Repository

```bash
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay
```

#### 2. 安装依赖 / Install Dependencies

```bash
# 下载 Go 模块依赖
# Download Go module dependencies
go mod download

# 验证依赖
# Verify dependencies
go mod verify
```

#### 3. 编译项目 / Build Project

```bash
# 使用 Make 编译 (推荐)
# Build using Make (recommended)
make build

# 或手动编译
# Or build manually
go build -o alimpay ./cmd/alimpay

# 编译所有平台版本
# Build for all platforms
make build-all
```

#### 4. 配置和启动 / Configure and Start

参考方式一的步骤 2-5 / Refer to Method 1 steps 2-5

---

## Docker部署 / Docker Deployment

### 方式一：使用官方镜像 / Method 1: Using Official Image

**即将推出 / Coming soon**

```bash
# 拉取镜像
# Pull image
docker pull chanhanzhan/alimpay:latest

# 运行容器
# Run container
docker run -d \
  --name alimpay \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/qrcode:/app/qrcode \
  chanhanzhan/alimpay:latest
```

### 方式二：自行构建镜像 / Method 2: Build Your Own Image

#### 1. 克隆代码 / Clone Code

```bash
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay
```

#### 2. 准备配置文件 / Prepare Configuration

```bash
# 复制并编辑配置文件
# Copy and edit configuration file
cp configs/config.example.yaml configs/config.yaml
vim configs/config.yaml

# 准备经营码图片
# Prepare business QR code image
cp /path/to/your/business_qr.png qrcode/business_qr.png
```

#### 3. 构建镜像 / Build Image

```bash
# 构建镜像
# Build image
docker build -t alimpay:latest .

# 或使用 Make
# Or use Make
make docker
```

#### 4. 运行容器 / Run Container

```bash
docker run -d \
  --name alimpay \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/qrcode:/app/qrcode \
  --restart unless-stopped \
  alimpay:latest
```

#### 5. 查看日志 / View Logs

```bash
# 查看容器日志
# View container logs
docker logs -f alimpay

# 查看应用日志
# View application logs
docker exec alimpay tail -f /app/logs/alimpay.log
```

---

## Docker Compose部署 / Docker Compose Deployment

### 1. 准备环境 / Prepare Environment

```bash
# 克隆代码
# Clone code
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay

# 准备配置文件
# Prepare configuration
cp configs/config.example.yaml configs/config.yaml
vim configs/config.yaml

# 准备经营码图片
# Prepare business QR code
cp /path/to/your/business_qr.png qrcode/business_qr.png
```

### 2. Docker Compose 配置 / Docker Compose Configuration

项目已包含 `docker-compose.yml` 文件，内容如下：

The project includes `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  alimpay:
    build: .
    container_name: alimpay
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml:ro
      - ./data:/app/data
      - ./logs:/app/logs
      - ./qrcode:/app/qrcode
    environment:
      - TZ=Asia/Shanghai
      - GIN_MODE=release
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

### 3. 启动服务 / Start Service

```bash
# 启动服务（后台运行）
# Start service (background)
docker-compose up -d

# 查看服务状态
# View service status
docker-compose ps

# 查看日志
# View logs
docker-compose logs -f

# 停止服务
# Stop service
docker-compose down

# 重启服务
# Restart service
docker-compose restart
```

---

## 生产环境部署 / Production Deployment

### 使用 Systemd 管理服务 / Using Systemd to Manage Service

#### 1. 创建 Systemd 服务文件 / Create Systemd Service File

```bash
sudo vim /etc/systemd/system/alimpay.service
```

**服务配置内容 / Service Configuration:**

```ini
[Unit]
Description=AliMPay Payment Gateway Service
Documentation=https://github.com/chanhanzhan/AliMPay
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/alimpay
ExecStart=/opt/alimpay/alimpay -config=/opt/alimpay/configs/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536

# 安全加固 / Security Hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/alimpay/data /opt/alimpay/logs

# 日志配置 / Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=alimpay

[Install]
WantedBy=multi-user.target
```

#### 2. 部署应用 / Deploy Application

```bash
# 创建部署目录
# Create deployment directory
sudo mkdir -p /opt/alimpay
sudo chown www-data:www-data /opt/alimpay

# 复制文件
# Copy files
sudo cp alimpay /opt/alimpay/
sudo cp -r configs /opt/alimpay/
sudo cp -r qrcode /opt/alimpay/
sudo mkdir -p /opt/alimpay/data /opt/alimpay/logs
sudo chown -R www-data:www-data /opt/alimpay
```

#### 3. 启动和管理服务 / Start and Manage Service

```bash
# 重载 systemd 配置
# Reload systemd configuration
sudo systemctl daemon-reload

# 启动服务
# Start service
sudo systemctl start alimpay

# 设置开机自启
# Enable auto-start on boot
sudo systemctl enable alimpay

# 查看服务状态
# Check service status
sudo systemctl status alimpay

# 查看日志
# View logs
sudo journalctl -u alimpay -f

# 停止服务
# Stop service
sudo systemctl stop alimpay

# 重启服务
# Restart service
sudo systemctl restart alimpay
```

---

## Nginx反向代理配置 / Nginx Reverse Proxy

### 基础配置 / Basic Configuration

```nginx
# /etc/nginx/sites-available/alimpay.conf

server {
    listen 80;
    server_name your-domain.com;  # 替换为你的域名 / Replace with your domain

    # 访问日志 / Access log
    access_log /var/log/nginx/alimpay_access.log;
    error_log /var/log/nginx/alimpay_error.log;

    # 客户端最大请求体大小 / Client max body size
    client_max_body_size 10M;

    # 代理到后端服务 / Proxy to backend service
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        
        # 传递真实IP / Pass real IP
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置 / Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # WebSocket 支持 (如需要) / WebSocket support (if needed)
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**启用配置 / Enable Configuration:**

```bash
# 创建软链接
# Create symbolic link
sudo ln -s /etc/nginx/sites-available/alimpay.conf /etc/nginx/sites-enabled/

# 测试配置
# Test configuration
sudo nginx -t

# 重载 Nginx
# Reload Nginx
sudo systemctl reload nginx
```

---

## HTTPS配置 / HTTPS Configuration

### 使用 Let's Encrypt 免费证书 / Using Let's Encrypt Free Certificate

#### 1. 安装 Certbot / Install Certbot

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install certbot python3-certbot-nginx
```

#### 2. 获取证书 / Obtain Certificate

```bash
# 自动配置 Nginx HTTPS
# Automatically configure Nginx HTTPS
sudo certbot --nginx -d your-domain.com

# 或者仅获取证书
# Or just obtain certificate
sudo certbot certonly --nginx -d your-domain.com
```

#### 3. Nginx HTTPS 配置 / Nginx HTTPS Configuration

```nginx
# /etc/nginx/sites-available/alimpay.conf

# HTTP 重定向到 HTTPS / HTTP redirect to HTTPS
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS 配置 / HTTPS configuration
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL 证书配置 / SSL certificate configuration
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/your-domain.com/chain.pem;

    # SSL 安全配置 / SSL security configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # HSTS (可选) / HSTS (optional)
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # 其他安全头 / Other security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # 访问日志 / Access log
    access_log /var/log/nginx/alimpay_access.log;
    error_log /var/log/nginx/alimpay_error.log;

    # 代理配置 / Proxy configuration
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

#### 4. 自动续期 / Auto-renewal

```bash
# 测试续期
# Test renewal
sudo certbot renew --dry-run

# Certbot 会自动设置 cron 任务进行续期
# Certbot automatically sets up a cron job for renewal
```

---

## 监控与维护 / Monitoring and Maintenance

### 健康检查 / Health Check

```bash
# 检查服务状态
# Check service status
curl http://localhost:8080/health

# 检查系统状态（更详细）
# Check system status (more details)
curl http://localhost:8080/health?action=status
```

### 日志管理 / Log Management

```bash
# 查看实时日志
# View real-time logs
tail -f logs/alimpay.log

# 查看最近的错误日志
# View recent error logs
grep "ERROR" logs/alimpay.log | tail -20

# 日志轮转（已通过配置自动进行）
# Log rotation (automatically done via configuration)
```

### 数据库备份 / Database Backup

```bash
# SQLite 数据库备份
# SQLite database backup
cp data/alimpay.db data/alimpay.db.backup.$(date +%Y%m%d)

# 自动备份脚本
# Automatic backup script
cat > /opt/alimpay/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/alimpay/backups"
mkdir -p $BACKUP_DIR
cp /opt/alimpay/data/alimpay.db $BACKUP_DIR/alimpay.db.$(date +%Y%m%d_%H%M%S)
# 保留最近 7 天的备份 / Keep last 7 days backups
find $BACKUP_DIR -name "alimpay.db.*" -mtime +7 -delete
EOF

chmod +x /opt/alimpay/backup.sh

# 添加到 crontab（每天凌晨2点备份）
# Add to crontab (backup at 2 AM daily)
# 0 2 * * * /opt/alimpay/backup.sh
```

### 性能监控 / Performance Monitoring

```bash
# 查看进程资源占用
# View process resource usage
ps aux | grep alimpay

# 查看端口监听
# View port listening
netstat -tlnp | grep 8080

# 或使用 ss
# Or use ss
ss -tlnp | grep 8080
```

---

## 常见问题 / Troubleshooting

### 问题 1：服务无法启动 / Service Won't Start

**排查步骤 / Troubleshooting Steps:**

```bash
# 1. 检查配置文件是否正确
# Check if configuration file is correct
./alimpay -config=./configs/config.yaml --check-config

# 2. 查看详细错误日志
# View detailed error logs
./alimpay -config=./configs/config.yaml --log-level=debug

# 3. 检查端口是否被占用
# Check if port is already in use
lsof -i :8080

# 4. 检查文件权限
# Check file permissions
ls -la configs/config.yaml
ls -la qrcode/business_qr.png
```

### 问题 2：无法访问服务 / Cannot Access Service

**排查步骤 / Troubleshooting Steps:**

```bash
# 1. 检查服务是否运行
# Check if service is running
systemctl status alimpay  # Systemd
docker ps  # Docker

# 2. 检查防火墙
# Check firewall
sudo ufw status  # Ubuntu
sudo firewall-cmd --list-all  # CentOS

# 3. 开放端口（如需要）
# Open port (if needed)
sudo ufw allow 8080  # Ubuntu
sudo firewall-cmd --add-port=8080/tcp --permanent  # CentOS
sudo firewall-cmd --reload

# 4. 检查 SELinux (CentOS/RHEL)
# Check SELinux (CentOS/RHEL)
sestatus
```

### 问题 3：支付回调失败 / Payment Callback Fails

**排查步骤 / Troubleshooting Steps:**

```bash
# 1. 检查回调地址是否可访问
# Check if callback URL is accessible
curl -I https://your-domain.com/notify

# 2. 查看回调日志
# View callback logs
grep "notify" logs/alimpay.log | tail -20

# 3. 检查签名验证
# Check signature verification
grep "signature" logs/alimpay.log | tail -20
```

### 问题 4：Docker 容器异常退出 / Docker Container Exits Abnormally

**排查步骤 / Troubleshooting Steps:**

```bash
# 1. 查看容器日志
# View container logs
docker logs alimpay

# 2. 检查容器状态
# Check container status
docker inspect alimpay

# 3. 进入容器排查
# Enter container for troubleshooting
docker exec -it alimpay /bin/sh

# 4. 查看健康检查
# View health check
docker inspect --format='{{.State.Health.Status}}' alimpay
```

---

## 性能优化建议 / Performance Optimization Tips

### 1. 数据库优化 / Database Optimization

```yaml
# configs/config.yaml
database:
  max_idle_conns: 20
  max_open_conns: 200
  conn_max_lifetime: 3600
```

### 2. 日志优化 / Logging Optimization

```yaml
# configs/config.yaml
logging:
  level: "warn"  # 生产环境使用 warn 或 error / Use warn or error in production
  compress: true  # 启用日志压缩 / Enable log compression
```

### 3. 监控优化 / Monitoring Optimization

```yaml
# configs/config.yaml
monitor:
  interval: 5  # 根据实际情况调整 / Adjust based on actual needs
```

---

## 安全加固建议 / Security Hardening Tips

1. **使用 HTTPS** / Use HTTPS
2. **定期更新系统和依赖** / Regularly update system and dependencies
3. **限制访问IP（如可能）** / Restrict access IPs (if possible)
4. **配置防火墙规则** / Configure firewall rules
5. **定期备份数据** / Regular data backups
6. **使用强密码** / Use strong passwords
7. **监控异常访问** / Monitor abnormal access
8. **及时更新应用版本** / Timely update application version

---

## 联系支持 / Contact Support

如有问题，请通过以下方式获取帮助：

If you have questions, get help through:

- **GitHub Issues**: https://github.com/chanhanzhan/AliMPay/issues
- **文档 / Documentation**: https://github.com/chanhanzhan/AliMPay/tree/main/docs
- **Email**: support@openel.top

---

**祝您部署顺利！/ Happy Deploying!** 🚀
