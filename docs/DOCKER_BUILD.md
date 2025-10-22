# Docker 镜像构建指南

本项目支持自动构建和发布 Docker 镜像到多个容器镜像仓库。

## 🚀 自动构建触发

镜像会在以下情况自动构建：

1. **推送到 main 分支** - 构建并发布 `latest` 标签
2. **创建 Git Tag** (如 `v1.0.0`) - 构建并发布版本标签
3. **Pull Request** - 仅构建测试，不发布
4. **手动触发** - 通过 GitHub Actions 界面手动运行

## 📦 支持的镜像仓库

### 1. GitHub Container Registry (GHCR)
**默认启用**，无需额外配置。

镜像地址：
```bash
ghcr.io/<username>/alimpay:latest
ghcr.io/<username>/alimpay:v1.0.0
```

拉取镜像：
```bash
docker pull ghcr.io/<username>/alimpay:latest
```

### 2. Docker Hub（可选）
需要配置 GitHub Secrets。

镜像地址：
```bash
docker.io/<dockerhub-username>/alimpay:latest
docker.io/<dockerhub-username>/alimpay:v1.0.0
```

拉取镜像：
```bash
docker pull <dockerhub-username>/alimpay:latest
```

## 🔧 配置 Docker Hub

### 步骤 1: 创建 Docker Hub Access Token

1. 登录 [Docker Hub](https://hub.docker.com/)
2. 进入 Account Settings → Security
3. 点击 "New Access Token"
4. 输入描述（如 "GitHub Actions"）
5. 选择权限：Read, Write, Delete
6. 复制生成的 token（只显示一次！）

### 步骤 2: 添加 GitHub Secrets

1. 进入 GitHub 仓库
2. 点击 Settings → Secrets and variables → Actions
3. 添加以下 secrets：

| Secret Name | 值 | 说明 |
|-------------|-----|------|
| `DOCKERHUB_USERNAME` | 你的 Docker Hub 用户名 | 必需 |
| `DOCKERHUB_TOKEN` | 你的 Access Token | 必需 |

## 🏷️ 镜像标签策略

工作流会自动生成以下标签：

### 基于 Git Tag
当你创建 tag 如 `v1.2.3` 时：
```
ghcr.io/user/alimpay:1.2.3
ghcr.io/user/alimpay:1.2
ghcr.io/user/alimpay:1
ghcr.io/user/alimpay:latest
```

### 基于分支
推送到 main 分支：
```
ghcr.io/user/alimpay:main
ghcr.io/user/alimpay:main-abc1234
ghcr.io/user/alimpay:latest
```

### Pull Request
```
ghcr.io/user/alimpay:pr-123
```

## 🖥️ 支持的平台

镜像支持多个 CPU 架构：
- `linux/amd64` - x86_64 (Intel/AMD)
- `linux/arm64` - ARM64 (Apple Silicon, Raspberry Pi 4+)

Docker 会自动拉取适合你系统的架构。

## 🔒 安全扫描

每次构建都会自动：
1. 使用 **Trivy** 扫描镜像漏洞
2. 将结果上传到 GitHub Security tab
3. 如发现高危漏洞，会在 Security 页面显示告警

查看扫描结果：
```
Repository → Security → Code scanning alerts
```

## 📊 构建摘要

每次成功构建后，GitHub Actions 会生成详细的摘要，包括：
- 发布的镜像标签
- 支持的平台
- 镜像 digest
- Pull 命令

## 🛠️ 手动触发构建

### 通过 GitHub 界面

1. 进入 Actions 标签
2. 选择 "Docker Build and Publish" 工作流
3. 点击 "Run workflow"
4. 选择分支
5. 点击 "Run workflow" 按钮

### 通过 Git Tag

```bash
# 创建并推送 tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 删除错误的 tag（如果需要）
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

## 🐳 本地构建

如果需要本地构建多平台镜像：

### 1. 设置 buildx
```bash
docker buildx create --use --name multiarch
docker buildx inspect --bootstrap
```

### 2. 构建多平台镜像
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag alimpay:latest \
  --build-arg VERSION=dev \
  --build-arg BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
  --push \
  .
```

### 3. 仅构建本地平台
```bash
docker build -t alimpay:latest .
```

## 📝 镜像信息

查看镜像详细信息：

```bash
# 查看镜像大小
docker images alimpay

# 查看镜像层
docker history alimpay:latest

# 查看镜像元数据
docker inspect alimpay:latest

# 查看支持的平台
docker buildx imagetools inspect ghcr.io/user/alimpay:latest
```

## 🔍 故障排查

### 构建失败

1. **查看日志**：Actions → 失败的工作流 → 点击查看详细日志
2. **常见问题**：
   - Dockerfile 语法错误
   - 依赖下载失败
   - 内存不足

### 推送失败

1. **GHCR 推送失败**：
   - 检查 `packages: write` 权限是否启用
   - 确认 GITHUB_TOKEN 有效

2. **Docker Hub 推送失败**：
   - 检查 secrets 是否正确配置
   - 确认 Access Token 未过期
   - 验证用户名和密码

### 无法拉取镜像

1. **GHCR 镜像**：
   ```bash
   # 公开仓库
   docker pull ghcr.io/user/alimpay:latest
   
   # 私有仓库需要先登录
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
   ```

2. **Docker Hub 镜像**：
   ```bash
   # 公开仓库
   docker pull username/alimpay:latest
   
   # 私有仓库
   docker login
   ```

## 🔄 工作流更新

如需修改构建流程，编辑：
```
.github/workflows/docker-publish.yml
```

主要配置：
- 触发条件：`on:` 部分
- 镜像仓库：`env:` 部分
- 构建参数：`build-args:` 部分
- 平台支持：`platforms:` 部分

## 📚 相关文档

- [GitHub Container Registry 文档](https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Docker Hub 文档](https://docs.docker.com/docker-hub/)
- [Docker Buildx 文档](https://docs.docker.com/buildx/working-with-buildx/)
- [Trivy 扫描器](https://github.com/aquasecurity/trivy)

## 💡 最佳实践

1. **语义化版本**：使用 `v1.2.3` 格式的 tag
2. **安全扫描**：定期检查 Security 标签的扫描结果
3. **镜像大小**：使用多阶段构建保持镜像小巧（当前 ~25MB）
4. **缓存策略**：利用 GitHub Actions cache 加速构建
5. **标签策略**：production 使用固定版本，development 使用 latest

## 🆘 获取帮助

遇到问题？
1. 查看 [GitHub Issues](https://github.com/user/alimpay/issues)
2. 阅读 [GitHub Actions 文档](https://docs.github.com/actions)
3. 查看工作流运行日志

---

**最后更新**: 2025-10-23  
**工作流版本**: v2.0

