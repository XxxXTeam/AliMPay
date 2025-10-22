# 贡献指南

感谢您对 AliMPay 项目的兴趣！我们欢迎各种形式的贡献。

## 行为准则

请阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。

## 如何贡献

### 报告Bug

在提交Bug报告之前，请先搜索现有的Issue，确保问题尚未被报告。

创建Bug报告时，请提供：

1. **清晰的标题**：简要描述问题
2. **详细描述**：详细说明问题
3. **复现步骤**：如何重现此问题
4. **预期行为**：您期望发生什么
5. **实际行为**：实际发生了什么
6. **环境信息**：OS、Go版本、项目版本等
7. **日志**：相关的日志信息
8. **截图**：如果适用

### 提出功能请求

我们欢迎新功能建议！提交功能请求时，请：

1. **检查现有Issue**：确保功能尚未被请求
2. **清晰描述**：说明您想要什么功能
3. **解释原因**：为什么需要这个功能
4. **提供示例**：如何使用这个功能
5. **考虑影响**：对现有功能的影响

### 提交代码

#### 开发流程

1. **Fork仓库**
   ```bash
   git clone https://github.com/chanhanzhan/alimpay.git
   cd alimpay-go
   ```

2. **创建分支**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **进行更改**
   - 编写代码
   - 添加测试
   - 更新文档

4. **提交更改**
   ```bash
   git add .
   git commit -m "Add some amazing feature"
   ```

5. **推送到分支**
   ```bash
   git push origin feature/amazing-feature
   ```

6. **创建Pull Request**

#### 代码规范

- **Go代码规范**：遵循[Effective Go](https://golang.org/doc/effective_go.html)
- **代码格式化**：使用`gofmt`和`goimports`
- **命名规范**：
  - 包名：小写，简短，有意义
  - 变量名：驼峰命名
  - 常量：全大写，下划线分隔
  - 函数：驼峰命名，导出函数首字母大写

- **注释规范**：
  - 所有导出的函数、类型和变量都应有注释
  - 注释应该说明"为什么"而不是"是什么"
  - 复杂逻辑添加详细注释

- **错误处理**：
  - 始终检查和处理错误
  - 使用`fmt.Errorf`包装错误
  - 提供有意义的错误消息

#### 代码质量检查

提交代码前，请确保通过以下检查：

```bash
# 格式化代码
make fmt

# 代码检查
make lint

# 运行测试
make test

# 安全检查
make security
```

#### 提交信息规范

使用清晰的提交信息，格式如下：

```
<类型>: <简短描述>

<详细描述>

<相关Issue>
```

类型：
- `feat`: 新功能
- `fix`: Bug修复
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

示例：
```
feat: 添加微信支付支持

- 实现微信支付API接口
- 添加微信支付配置
- 更新文档说明

Closes #123
```

### Pull Request流程

1. **PR描述**：
   - 清晰描述变更内容
   - 说明变更原因
   - 列出相关Issue

2. **代码审查**：
   - 等待维护者审查
   - 根据反馈进行修改
   - 保持PR简洁，专注单一功能

3. **CI检查**：
   - 确保所有CI检查通过
   - 修复失败的测试
   - 解决代码质量问题

4. **合并**：
   - 维护者批准后合并
   - 使用Squash合并保持历史清晰

## 开发环境设置

### 前置要求

- Go 1.23+
- Git
- Make
- Docker（可选）

### 设置步骤

1. **克隆仓库**
   ```bash
   git clone https://github.com/chanhanzhan/alimpay-go.git
   cd alimpay-go
   ```

2. **安装依赖**
   ```bash
   make install
   ```

3. **配置项目**
   ```bash
   cp configs/config.example.yaml configs/config.yaml
   # 编辑配置文件
   ```

4. **初始化数据库**
   ```bash
   make init
   ```

5. **运行项目**
   ```bash
   make dev
   ```

### 开发工具

推荐使用以下工具：

- **IDE**: GoLand, VSCode with Go extension
- **代码检查**: golangci-lint
- **调试**: Delve
- **API测试**: Postman, curl
- **Docker**: Docker Desktop

## 测试

### 运行测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage

# 运行基准测试
make bench
```

### 编写测试

- 每个包都应该有测试文件
- 测试文件命名：`*_test.go`
- 测试函数命名：`TestXxx`
- 使用表驱动测试
- 测试覆盖率应 > 70%

示例：
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "output2", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 文档

### 更新文档

如果您的更改影响用户界面或API，请更新相关文档：

- `README.md`: 项目概述和快速开始
- `docs/API.md`: API文档
- `CHANGELOG.md`: 更新日志
- 代码注释：GoDoc格式

### 文档规范

- 使用Markdown格式
- 保持语言简洁明了
- 提供代码示例
- 添加必要的截图

## 版本发布

维护者负责版本发布，流程如下：

1. 更新版本号
2. 更新CHANGELOG
3. 创建Git标签
4. 推送标签触发自动发布
5. 编写Release Notes

## 社区

- **GitHub Issues**: 问题讨论和Bug报告
- **Pull Requests**: 代码贡献
- **Discussions**: 功能讨论和问答

## 许可证

通过贡献代码，您同意您的贡献将在[MIT许可证](LICENSE)下授权。

## 问题？

如有任何问题，请：

1. 查看[FAQ](docs/FAQ.md)
2. 搜索现有Issue
3. 创建新Issue
4. 发送邮件至：support@alimpay.com

---

再次感谢您的贡献！您的支持使这个项目变得更好。

**Happy Coding! 🚀**

