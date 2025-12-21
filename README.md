# AliMPay Golang Edition

<div align="center">

[![Build Status](https://github.com/alimpay/alimpay-go/workflows/Build%20and%20Test/badge.svg)](https://github.com/alimpay/alimpay-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/alimpay/alimpay-go)](https://goreportcard.com/report/github.com/alimpay/alimpay-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://golang.org/dl/)

高性能支付宝码支付/易支付接口系统 Golang 实现

[功能特性](#功能特性) • [快速开始](#快速开始) • [文档](#文档) • [API文档](#api文档) • [部署指南](#部署指南) • [常见问题](#常见问题)

</div>

---

## 📖 简介

AliMPay Golang Edition 是一个基于 Go 语言开发的高性能支付宝码支付系统，完全兼容易支付和码支付标准接口，支持经营码收款和动态转账两种模式。

### ✨ 功能特性

- 🚀 **高性能**: 基于 Go 和 Gin 框架，高并发处理能力
- 💳 **多支付模式**: 
  - 经营码收款模式（推荐）
  - 动态转账二维码模式
  - **多二维码轮询模式** ⭐ 支持负载均衡
- 🏢 **多商户支持**: 
  - **每个二维码独立API配置** ⭐ NEW
  - 支持多个支付宝商户账号
  - 业务线级别隔离
- 🔒 **安全可靠**: 
  - RSA2 签名验证
  - 防0元购保护
  - SQL注入防护
  - XSS防护
- 🎯 **标准接口**: 完全兼容易支付和码支付API
- 📊 **管理后台**: 现代化的订单管理界面
- 🔄 **自动监听**: 账单查询自动匹配支付
- 🔀 **智能轮询**: 支持多二维码轮询，提高并发处理能力
- 📦 **独立部署**: 无需PHP环境，一键部署
- 🐳 **容器化**: 支持Docker镜像快速部署
- 📈 **实时监控**: 订单状态实时查询和更新

### 🏗️ 技术栈

- **后端**: Go 1.23+, Gin Web Framework
- **数据库**: SQLite3 (可扩展为MySQL/PostgreSQL)
- **缓存**: Redis (可选)
- **日志**: Zap
- **定时任务**: Cron
- **前端**: 原生 JavaScript + Modern CSS

---

## 📚 文档 / Documentation

**📖 [完整文档中心](docs/README.md)** - 查看所有文档的索引和导航

完整的文档帮助您快速上手和深入了解系统：

### 新手入门 / Getting Started
- **[📖 快速开始指南](docs/QUICKSTART.md)** - 10分钟快速部署运行
- **[🚀 部署教程](docs/DEPLOYMENT.md)** - 详细的部署指南（Docker、Systemd、Nginx等）
- **[🔌 接入教程](docs/INTEGRATION.md)** - 如何集成到您的应用（含多语言示例）

### 参考文档 / Reference
- **[📡 API 文档](docs/API.md)** - 完整的 API 接口说明
- **[🔀 多二维码轮询](docs/MULTI_QRCODE.md)** - 多二维码轮询功能详解
- **[🏢 多二维码独立API](docs/MULTI_QR_API.md)** - 每个二维码使用独立支付宝API配置 ⭐ NEW
- **[❓ 常见问题](docs/FAQ.md)** - 常见问题解答
- **[⚙️ 配置说明](configs/config.example.yaml)** - 详细的配置文件注释
- **[🔧 易支付兼容性](EPAY_COMPATIBILITY.md)** - 易支付/码支付兼容说明

### 贡献 / Contributing
- **[🤝 贡献指南](CONTRIBUTING.md)** - 如何参与项目贡献
- **[📝 提交规范](docs/COMMIT_GUIDELINES.md)** - Git 提交信息规范（即将添加）

---

## 🚀 快速开始

### 方式一：使用 Docker 镜像（推荐） 🐳

**最简单的方式，无需编译，开箱即用！**

#### 1. 准备配置文件

```bash
# 创建工作目录
mkdir -p alimpay/{configs,data,logs,qrcode}
cd alimpay

# 下载配置文件模板
wget https://raw.githubusercontent.com/chanhanzhan/AliMPay/main/configs/config.example.yaml -O configs/config.yaml

# 编辑配置文件
vim configs/config.yaml
```

配置必需项：

```yaml
alipay:
  app_id: "你的支付宝应用ID"                    # 从支付宝开放平台获取
  private_key: "你的应用私钥"                   # 使用密钥生成工具生成
  alipay_public_key: "支付宝公钥"               # 从支付宝开放平台获取
  transfer_user_id: "收款支付宝用户ID"          # 您的支付宝账号UID

payment:
  business_qr_mode:
    enabled: true                               # 启用经营码模式（推荐）
    qr_code_path: "./qrcode/business_qr.png"   # 经营码图片路径
```

#### 2. 放置收款二维码

```bash
# 将您的支付宝经营码图片放到 qrcode 目录
cp your_qrcode.png qrcode/business_qr.png
```

#### 3. 拉取并运行镜像

```bash
# 从 GitHub Container Registry 拉取（推荐）
docker pull ghcr.io/chanhanzhan/alimpay:latest

# 或从 Docker Hub 拉取
docker pull chanhanzhan/alimpay:latest

# 运行容器
docker run -d \
  --name alimpay \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/qrcode:/app/qrcode:ro \
  --restart unless-stopped \
  ghcr.io/chanhanzhan/alimpay:latest
```

#### 4. 访问系统

- **支付接口**: http://your-server-ip:8080/submit
- **管理后台**: http://your-server-ip:8080/admin/dashboard
- **健康检查**: http://your-server-ip:8080/health

**查看日志**:
```bash
docker logs -f alimpay
```

**停止服务**:
```bash
docker stop alimpay
docker rm alimpay
```

---

### 方式二：使用 Docker Compose（推荐用于生产环境）

Docker Compose 提供了更完整的部署方案，支持健康检查、日志管理等功能。

#### 步骤 1：准备项目文件

```bash
# 创建项目目录
mkdir -p alimpay && cd alimpay

# 下载 docker-compose.yml
wget https://raw.githubusercontent.com/chanhanzhan/AliMPay/main/docker-compose.yml

# 创建必要的目录
mkdir -p configs data logs qrcode

# 下载配置文件模板
wget https://raw.githubusercontent.com/chanhanzhan/AliMPay/main/configs/config.example.yaml -O configs/config.yaml

# 编辑配置
vim configs/config.yaml
```

#### 步骤 2：配置文件说明

项目的 `docker-compose.yml` 包含以下特性：

```yaml
version: '3.8'

services:
  alimpay:
    image: ghcr.io/chanhanzhan/alimpay:latest
    container_name: alimpay
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml:ro
      - ./data:/app/data
      - ./logs:/app/logs
      - ./qrcode:/app/qrcode
    environment:
      - TZ=Asia/Shanghai              # 时区设置
      - GIN_MODE=release              # 生产模式
    healthcheck:                       # 健康检查
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
    logging:                           # 日志管理
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

#### 步骤 3：启动服务

```bash
# 拉取最新镜像
docker-compose pull

# 启动服务（后台运行）
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f alimpay

# 查看最近100行日志
docker-compose logs --tail=100 alimpay
```

#### 步骤 4：管理服务

```bash
# 停止服务
docker-compose stop

# 启动服务
docker-compose start

# 重启服务
docker-compose restart

# 停止并删除容器
docker-compose down

# 停止并删除容器及数据卷
docker-compose down -v
```

#### 可选：启用 Redis 缓存

项目支持可选的 Redis 缓存服务：

```bash
# 使用 Redis profile 启动
docker-compose --profile with-redis up -d

# 查看 Redis 状态
docker-compose ps redis
```

#### 升级到新版本

```bash
# 拉取新镜像
docker-compose pull

# 重启服务
docker-compose up -d

# 查看日志确认启动成功
docker-compose logs -f
```

---

### 方式三：本地编译部署

适合需要自定义修改或开发的用户。

#### 环境要求

| 依赖 | 版本要求 | 用途 | 安装检查 |
|------|----------|------|----------|
| **Go** | 1.23+ | 编译和运行 | `go version` |
| **Git** | 2.0+ | 克隆代码 | `git --version` |
| **Make** | 3.8+ | 构建工具 | `make --version` |
| **GCC** | 可选 | CGO 编译 SQLite | `gcc --version` |

#### 步骤 1：安装依赖

<details>
<summary><b>Linux (Ubuntu/Debian)</b></summary>

```bash
# 更新软件包列表
sudo apt update

# 安装 Go（如未安装）
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# 安装其他依赖
sudo apt install -y git make gcc

# 验证安装
go version
git --version
make --version
```

</details>

<details>
<summary><b>Linux (CentOS/RHEL)</b></summary>

```bash
# 安装 Go
sudo yum install -y golang

# 或手动安装最新版本
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 安装其他依赖
sudo yum install -y git make gcc

# 验证安装
go version
```

</details>

<details>
<summary><b>macOS</b></summary>

```bash
# 使用 Homebrew 安装（推荐）
brew install go git

# 或下载安装包
# 访问 https://go.dev/dl/ 下载 macOS 安装包

# 验证安装
go version
git --version
make --version  # macOS 自带 make
```

</details>

<details>
<summary><b>Windows</b></summary>

```powershell
# 1. 下载 Go 安装包
# 访问 https://go.dev/dl/ 下载 Windows 安装包并安装

# 2. 安装 Git
# 访问 https://git-scm.com/download/win 下载并安装

# 3. 安装 Make（可选）
# 下载 GnuWin32 Make: http://gnuwin32.sourceforge.net/packages/make.htm
# 或使用 Chocolatey: choco install make

# 4. 验证安装
go version
git --version
make --version
```

</details>

#### 步骤 2：配置 Go 环境

```bash
# 配置 Go 模块代理（加速依赖下载）
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 查看 Go 环境配置
go env
```

#### 步骤 3：克隆代码

```bash
# 克隆仓库
git clone https://github.com/chanhanzhan/AliMPay.git
cd AliMPay

# 查看项目结构
tree -L 2  # 或 ls -la
```

#### 步骤 4：安装项目依赖

```bash
# 下载 Go 模块依赖
go mod download

# 验证依赖完整性
go mod verify

# 查看依赖列表
go list -m all
```

#### 步骤 5：配置应用

```bash
# 复制配置文件模板
cp configs/config.example.yaml configs/config.yaml

# 编辑配置（填写支付宝API信息）
vim configs/config.yaml
# 或使用其他编辑器：nano、code、gedit 等

# 准备二维码目录
mkdir -p qrcode
# 将您的支付宝收款码图片放到 qrcode/ 目录
```

#### 步骤 6：编译和运行

##### 方式 A：使用 Make（推荐）

```bash
# 查看可用命令
make help

# 开发模式运行（自动重启）
make dev

# 构建生产版本
make build

# 运行编译后的程序
./alimpay -config=./configs/config.yaml

# 其他有用的命令
make test          # 运行测试
make lint          # 代码检查
make clean         # 清理编译文件
```

##### 方式 B：直接使用 Go 命令

```bash
# 开发模式运行
go run ./cmd/alimpay -config=./configs/config.yaml

# 编译
go build -o alimpay ./cmd/alimpay

# 运行
./alimpay -config=./configs/config.yaml

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -o alimpay-linux-amd64 ./cmd/alimpay

# 交叉编译（Windows）
GOOS=windows GOARCH=amd64 go build -o alimpay-windows-amd64.exe ./cmd/alimpay

# 交叉编译（macOS）
GOOS=darwin GOARCH=amd64 go build -o alimpay-darwin-amd64 ./cmd/alimpay
```

#### 步骤 7：验证运行

```bash
# 访问健康检查接口
curl http://localhost:8080/health

# 查看日志
tail -f logs/alimpay.log

# 访问管理后台
open http://localhost:8080/admin/dashboard
# 或在浏览器中打开 http://localhost:8080/admin/dashboard
```

#### 开发工具推荐

- **IDE**: 
  - [GoLand](https://www.jetbrains.com/go/) - JetBrains 专业 Go IDE
  - [VS Code](https://code.visualstudio.com/) + [Go 插件](https://marketplace.visualstudio.com/items?itemName=golang.go)
  
- **调试工具**:
  - [Delve](https://github.com/go-delve/delve) - Go 调试器

- **代码检查**:
  - [golangci-lint](https://golangci-lint.run/) - 代码质量检查

#### 常见问题

<details>
<summary>依赖下载失败？</summary>

```bash
# 尝试使用国内镜像
go env -w GOPROXY=https://goproxy.cn,direct

# 或使用阿里云镜像
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
```

</details>

<details>
<summary>CGO 相关错误？</summary>

```bash
# 如果不需要 CGO，可以禁用
CGO_ENABLED=0 go build ./cmd/alimpay

# 或在 Linux 上安装 GCC
sudo apt install build-essential  # Ubuntu/Debian
sudo yum groupinstall "Development Tools"  # CentOS/RHEL
```

</details>

<details>
<summary>端口被占用？</summary>

```bash
# Linux/macOS
sudo lsof -i :8080
sudo kill -9 <PID>

# 或修改配置文件中的端口
vim configs/config.yaml
# server.port: 8080 -> server.port: 8081
```

</details>

**详细开发指南：** [📖 快速开始指南](docs/QUICKSTART.md) | [🤝 贡献指南](CONTRIBUTING.md)

---

## 🐳 Docker 镜像源

### 官方镜像仓库

| 镜像源 | 拉取命令 | 说明 |
|--------|----------|------|
| **GitHub Container Registry (GHCR)** | `docker pull ghcr.io/chanhanzhan/alimpay:latest` | 官方镜像仓库 ⭐ |

> 💡 **提示：** 我们使用 GitHub Container Registry 作为官方镜像仓库，提供稳定可靠的镜像服务。

### 可用标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `latest` | 最新稳定版（main 分支） | `ghcr.io/chanhanzhan/alimpay:latest` |
| `v{version}` | 指定版本号 | `ghcr.io/chanhanzhan/alimpay:v1.1.0` |
| `v{major}.{minor}` | 主次版本号 | `ghcr.io/chanhanzhan/alimpay:v1.1` |
| `v{major}` | 主版本号 | `ghcr.io/chanhanzhan/alimpay:v1` |
| `{branch}-{sha}` | 分支+提交SHA | `ghcr.io/chanhanzhan/alimpay:main-abc123` |

### 镜像架构支持

- ✅ **linux/amd64** - x86_64 架构（常见服务器）
- ✅ **linux/arm64** - ARM64 架构（树莓派、ARM 服务器）

Docker 会自动选择与您系统匹配的架构。

### 镜像信息

```bash
# 查看镜像详细信息
docker image inspect ghcr.io/chanhanzhan/alimpay:latest

# 查看镜像架构
docker manifest inspect ghcr.io/chanhanzhan/alimpay:latest

# 查看本地镜像
docker images | grep alimpay

# 拉取指定架构的镜像
docker pull --platform linux/amd64 ghcr.io/chanhanzhan/alimpay:latest
```

**详细部署教程：** [🚀 部署指南](docs/DEPLOYMENT.md)

---

## 📡 API 文档

AliMPay 完全兼容易支付和码支付标准接口。

**完整 API 文档：** [📡 API Reference](docs/API.md)  
**接入教程：** [🔌 集成指南](docs/INTEGRATION.md)

### 码支付标准接口

#### 1. 创建订单

**请求地址**: `/submit` 或 `/api/submit`

**请求方式**: `GET` / `POST`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| type | string | 是 | 支付方式（alipay） |
| out_trade_no | string | 是 | 商户订单号 |
| notify_url | string | 是 | 异步通知地址 |
| return_url | string | 是 | 同步返回地址 |
| name | string | 是 | 商品名称 |
| money | string | 是 | 订单金额 |
| sign | string | 是 | 签名 |
| sign_type | string | 否 | 签名类型(默认MD5) |

**签名规则**:

```
1. 将所有参数(除sign和sign_type)按参数名ASCII码升序排列
2. 拼接成: key1=value1&key2=value2&key3=value3
3. 在末尾追加商户密钥: key1=value1&key2=value2{merchant_key}
4. MD5加密后转小写
```

**响应示例**:

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "trade_no": "20240115120000123456",
  "out_trade_no": "ORDER20240115001",
  "money": "1.00",
  "payment_amount": 1.01,
  "payment_url": "http://your-domain.com/pay?trade_no=xxx&amount=1.01",
  "qr_code": "data:image/png;base64,..."
}
```

#### 2. 查询订单

**请求地址**: `/api/order` 或 `/mapi?act=order`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| out_trade_no | string | 是 | 商户订单号 |

**响应示例**:

```json
{
  "code": 1,
  "msg": "SUCCESS",
  "trade_no": "20240115120000123456",
  "out_trade_no": "ORDER20240115001",
  "type": "alipay",
  "name": "测试商品",
  "money": "1.00",
  "status": 1,
  "addtime": "2024-01-15 12:00:00",
  "endtime": "2024-01-15 12:01:00"
}
```

**状态说明**:
- `0`: 待支付
- `1`: 已支付
- `2`: 已关闭

#### 3. 查询商户信息

**请求地址**: `/api?action=query`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |

#### 4. 关闭订单

**请求地址**: `/api/close`

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| pid | string | 是 | 商户ID |
| key | string | 是 | 商户密钥 |
| out_trade_no | string | 是 | 商户订单号 |

### 易支付兼容接口

系统完全兼容易支付接口标准，可以无缝替换易支付系统。

---

## 🛠️ 开发指南

### 项目结构

```
AliMPay/
├── cmd/alimpay/           # 主程序入口
├── configs/               # 配置文件
├── data/                  # 数据目录
├── logs/                  # 日志目录
├── qrcode/                # 二维码目录
└── internal/              # 内部包（所有核心代码）
    ├── config/            # 配置管理
    ├── database/          # 数据库操作
    ├── events/            # 事件系统
    ├── handler/           # HTTP处理器
    ├── middleware/        # 中间件
    ├── model/             # 数据模型
    ├── pkg/               # 工具包
    │   ├── cache/         # 缓存
    │   ├── lock/          # 锁机制
    │   ├── logger/        # 日志
    │   ├── qrcode/        # 二维码生成
    │   └── utils/         # 工具函数
    ├── response/          # 响应处理
    ├── scripts/           # 脚本工具
    ├── service/           # 业务逻辑
    ├── validator/         # 参数验证
    ├── web/               # 前端资源
    │   ├── static/        # 静态文件
    │   └── templates/     # HTML模板
    └── worker/            # 工作池
```

### Make 命令

```bash
# 查看所有命令
make help

# 构建
make build              # 编译项目
make build-all          # 编译所有平台版本
make release            # 创建发布版本

# 运行
make run                # 运行项目
make dev                # 开发模式运行

# 数据库
make init               # 初始化数据库
make db-reset           # 重置数据库

# 测试
make test               # 运行测试
make test-coverage      # 生成覆盖率报告
make bench              # 运行基准测试

# 代码质量
make fmt                # 格式化代码
make lint               # 代码检查
make security           # 安全检查
make tidy               # 整理依赖

# 工具
make clean              # 清理编译文件
make clean-all          # 深度清理
make docker             # 构建Docker镜像
```

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| GIN_MODE | 运行模式 | release |
| TZ | 时区 | Asia/Shanghai |

---

## 📊 监控与日志

### 健康检查

```bash
# 系统状态
curl http://localhost:8080/health?action=status

# 触发监控
curl http://localhost:8080/health?action=monitor

# 清理过期订单
curl http://localhost:8080/health?action=cleanup
```

### 日志查看

```bash
# 查看实时日志
tail -f logs/alimpay.log

# 使用Docker查看
docker-compose logs -f alimpay
```

---

## 🔧 常见问题

### Q: 如何获取支付宝相关配置？

**A:** 详细步骤请查看 [快速开始指南](docs/QUICKSTART.md#步骤-1-获取支付宝配置--step-1-get-alipay-configuration)

简要步骤：
1. 登录支付宝开放平台：https://open.alipay.com
2. 创建应用并获取 `app_id`
3. 使用密钥生成工具生成应用私钥和公钥
4. 上传应用公钥并获取支付宝公钥
5. 在账号中心查看账号UID（用户ID）
6. 开通相关接口权限

### Q: 经营码收款和转账模式有什么区别？

**A:**
- **经营码模式**（推荐）：使用固定的经营码收款，系统通过金额匹配订单，到账快，用户体验好
- **转账模式**：动态生成转账二维码，每个订单独立二维码，更灵活但配置相对复杂

推荐使用经营码模式，只需上传一张经营码图片即可。

### Q: 如何查看商户ID和密钥？

**A:** 首次运行后会自动生成，有以下几种查看方式：

1. **查看日志**
   ```bash
   tail -f logs/alimpay.log | grep "Merchant"
   ```

2. **查看配置文件**
   ```bash
   cat configs/config.yaml | grep -A 2 "merchant:"
   ```

3. **通过API查询**
   ```bash
   curl "http://localhost:8080/api?action=query&pid=YOUR_PID&key=YOUR_KEY"
   ```

### Q: 支付后没有自动跳转？

**A:** 请检查以下几点：

1. **监控服务是否启用**
   ```yaml
   monitor:
     enabled: true  # 必须为 true
   ```

2. **支付宝API权限**
   - 确认已开通"查询对账单下载地址"权限
   - 确认支付宝配置正确

3. **查看日志**
   ```bash
   tail -f logs/alimpay.log | grep "monitor"
   ```

4. **手动触发监控**
   ```bash
   curl "http://localhost:8080/health?action=monitor"
   ```

**更多问题请查看：** [❓ 常见问题文档](docs/FAQ.md)

---

## 🆕 新功能亮点

### 多二维码独立API配置 ⭐ NEW

**现在每个二维码可以配置独立的支付宝API！**

这意味着您可以：
- 🏢 使用多个支付宝商户账号
- 💼 不同业务线使用不同的支付账号
- ⚖️ 分散支付流量，降低单账号风险
- 🛡️ 实现账号级别的业务隔离

**配置示例**：

```yaml
payment:
  business_qr_mode:
    enabled: true
    qr_code_paths:
      # 商户A - 使用独立API
      - id: "merchant_a"
        path: "./qrcode/qr_a.png"
        code_id: "fkx111111"
        enabled: true
        priority: 1
        alipay_api:                    # ⭐ 独立API配置
          app_id: "2021001111111111"
          private_key: "..."
          alipay_public_key: "..."
          transfer_user_id: "2088111111111111"
      
      # 商户B - 使用全局配置
      - id: "merchant_b"
        path: "./qrcode/qr_b.png"
        code_id: "fkx222222"
        enabled: true
        priority: 2
        # 不配置 alipay_api，使用全局配置
```

**特性**：
- ✅ 智能配置合并（缺失字段自动补充）
- ✅ 自动服务创建（启动时自动识别）
- ✅ 订单级别匹配（每个订单使用对应API）
- ✅ 向后兼容（现有配置无需修改）

**详细文档**：
- [🏢 多二维码独立API配置指南](docs/MULTI_QR_API.md) - 完整的配置说明和使用案例
- [✨ 功能特性说明](FEATURE_MULTI_API.md) - 快速了解新功能
- [📋 更新日志](CHANGELOG_MULTI_API.md) - 详细的技术实现

---

## 📝 更新日志

### v1.1.0 (2024-10-24) 🎉

**新增功能**：
- ✨ **多二维码独立API配置** - 每个二维码可使用独立的支付宝API
- 🏢 **多商户账号支持** - 支持多个支付宝商户账号同时运行
- 🔍 **智能配置合并** - 自动合并全局和独立配置
- 📊 **订单级API匹配** - 每个订单自动使用对应的API查询

**功能增强**：
- 🚀 监控服务支持多API账单查询
- 📈 订单监听任务支持降级容错
- 📖 完善的配置文档和示例

**配置文件**：
- 新增 `configs/config.multi_api.example.yaml` - 多API配置示例
- 新增 `docs/MULTI_QR_API.md` - 详细配置指南
- 更新 `configs/config.example.yaml` - 添加独立API配置说明

### v1.0.0 (2024-01-15)

- ✨ 初始版本发布
- 🎉 支持经营码收款模式
- 🚀 完整的易支付接口实现
- 💎 现代化管理后台
- 🐳 Docker支持
- 🔄 自动监听支付状态
- 📊 实时订单统计

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 贡献流程

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改（遵循提交规范）
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 📝 提交规范

本项目采用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```bash
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响逻辑）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `build`: 构建系统或依赖变更
- `ci`: CI 配置变更
- `chore`: 其他变更

**示例**：
```bash
feat(api): add payment callback endpoint
fix(database): prevent deadlock in order query
docs: update README with Docker instructions
perf(logger): reduce memory allocation
```

详细规范请参考 [提交指南](docs/COMMIT_GUIDELINES.md)

### 设置提交模板

```bash
git config commit.template .gitmessage
```

更多贡献指南请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📜 开源协议

本项目采用 MIT 协议开源，详见 [LICENSE](LICENSE) 文件。

---

## 💖 致谢

- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [Zap](https://github.com/uber-go/zap) - 日志库
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite驱动
- [go-qrcode](https://github.com/skip2/go-qrcode) - 二维码生成

---

## 📧 联系方式

- Issue: https://github.com/chanhanzhan/alimpay/issues
- Email: support@openel.top

---

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐️ Star 支持一下！**

Made with ❤️ by AliMPay Team

</div>
