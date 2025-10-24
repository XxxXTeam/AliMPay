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
  - 多二维码轮询模式（新增）
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
- 🐳 **容器化**: 支持Docker一键部署
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
- **[❓ 常见问题](docs/FAQ.md)** - 常见问题解答
- **[⚙️ 配置说明](configs/config.example.yaml)** - 详细的配置文件注释
- **[🔧 易支付兼容性](EPAY_COMPATIBILITY.md)** - 易支付/码支付兼容说明

### 贡献 / Contributing
- **[🤝 贡献指南](CONTRIBUTING.md)** - 如何参与项目贡献
- **[📝 提交规范](docs/COMMIT_GUIDELINES.md)** - Git 提交信息规范（即将添加）

---

## 🚀 快速开始

### 一键体验 / Quick Experience

**只需三步即可开始使用：**

1. **准备支付宝配置** - 从支付宝开放平台获取 AppID 和密钥
2. **部署 AliMPay** - 使用 Docker 或直接运行
3. **开始接收支付** - 集成 API 到您的应用

**详细步骤请查看：** [📖 快速开始指南](docs/QUICKSTART.md)

### 环境要求

- Go 1.23 或更高版本
- Git (用于克隆代码)

### 安装步骤

#### 1. 克隆代码

```bash
git clone https://github.com/chanhanzhan/alimpay.git
cd alimpay-go
```

#### 2. 配置文件

```bash
# 复制配置文件模板
cp configs/config.example.yaml configs/config.yaml

# 编辑配置文件，填写支付宝相关信息
vim configs/config.yaml
```

**必需配置项：**

详细的配置说明请查看 [配置文件注释](configs/config.example.yaml)

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
    qr_code_id: ""                              # 可选：收款码ID，用于拉起支付宝
```

> 💡 **提示：** 配置文件包含详细的中英文注释，每个配置项都有说明和示例。

#### 3. 初始化数据库

```bash
make init
```

#### 4. 编译运行

```bash
# 开发模式运行
make dev

# 或编译后运行
make build
./alimpay -config=./configs/config.yaml
```

#### 5. 访问系统

- **支付接口**: http://localhost:8080/submit
- **管理后台**: http://localhost:8080/admin/dashboard
- **健康检查**: http://localhost:8080/health

---

## 🐳 Docker 部署

Docker 是最简单的部署方式，推荐生产环境使用。

**详细部署教程：** [🚀 部署指南](docs/DEPLOYMENT.md)

### 使用 Docker

```bash
# 构建镜像
docker build -t alimpay:latest .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/qrcode:/app/qrcode \
  --name alimpay \
  alimpay:latest
```

### 使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

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
alimpay-go/
├── cmd/alimpay/          # 主程序入口
├── internal/             # 内部包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库操作
│   ├── handler/         # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   └── validator/       # 参数验证
├── pkg/                 # 公共包
│   ├── cache/          # 缓存
│   ├── lock/           # 锁机制
│   ├── logger/         # 日志
│   ├── qrcode/         # 二维码生成
│   └── utils/          # 工具函数
├── web/                # 前端资源
│   ├── static/         # 静态文件
│   └── templates/      # HTML模板
├── configs/            # 配置文件
├── data/               # 数据目录
├── logs/               # 日志目录
└── qrcode/             # 二维码目录
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

## 📝 更新日志

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
