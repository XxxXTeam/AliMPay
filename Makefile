.PHONY: build run clean test init help install fmt lint dev docker release deps security

# 变量定义
BINARY_NAME=alimpay
CONFIG_PATH=./configs/config.yaml
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
help:
	@echo "AliMPay Golang Version - Makefile"
	@echo ""
	@echo "构建命令:"
	@echo "  make build         - 编译项目"
	@echo "  make build-all     - 编译所有平台版本"
	@echo "  make release       - 创建发布版本"
	@echo ""
	@echo "运行命令:"
	@echo "  make run           - 运行项目"
	@echo "  make dev           - 开发模式运行"
	@echo ""
	@echo "数据库命令:"
	@echo "  make init          - 初始化数据库"
	@echo "  make db-reset      - 重置数据库"
	@echo ""
	@echo "测试命令:"
	@echo "  make test          - 运行测试"
	@echo "  make test-coverage - 运行测试并生成覆盖率报告"
	@echo "  make bench         - 运行基准测试"
	@echo ""
	@echo "代码质量:"
	@echo "  make fmt           - 格式化代码"
	@echo "  make lint          - 代码检查"
	@echo "  make security      - 安全检查"
	@echo "  make tidy          - 整理依赖"
	@echo ""
	@echo "工具命令:"
	@echo "  make install       - 安装依赖"
	@echo "  make deps          - 更新依赖"
	@echo "  make clean         - 清理编译文件"
	@echo "  make docker        - 构建Docker镜像"
	@echo ""

# 安装依赖
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 编译项目
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/alimpay
	@echo "Build complete: $(BINARY_NAME) $(VERSION)"

# 编译所有平台版本
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/alimpay
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/alimpay
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/alimpay
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/alimpay
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/alimpay
	@echo "Build complete for all platforms"

# 创建发布版本
release: build-all
	@echo "Creating release archives..."
	cd dist && \
	tar czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64 && \
	tar czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release archives created in dist/"

# 运行项目
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME) -config=$(CONFIG_PATH)

# 初始化数据库
init:
	@echo "Initializing database..."
	go run scripts/init_db.go -config=$(CONFIG_PATH)

# 重置数据库
db-reset:
	@echo "Resetting database..."
	rm -rf data/alimpay.db
	go run scripts/init_db.go -config=$(CONFIG_PATH)
	@echo "Database reset complete"

# 清理编译文件
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	@echo "Clean complete"

# 深度清理（包括数据和日志）
clean-all: clean
	@echo "Deep cleaning..."
	rm -rf data/
	rm -rf logs/
	@echo "Deep clean complete"

# 运行测试
test:
	@echo "Running tests..."
	go test -v -race ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 基准测试
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# 代码格式化
fmt:
	@echo "Formatting code..."
	gofmt -w -s .
	goimports -w .

# 代码检查
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# 安全检查
security:
	@echo "Running security checks..."
	gosec ./...
	go list -json -m all | nancy sleuth

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# 更新依赖
deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 开发模式运行
dev:
	@echo "Running in development mode..."
	GIN_MODE=debug go run ./cmd/alimpay -config=$(CONFIG_PATH)

# Docker构建
docker:
	@echo "Building Docker image..."
	docker build -t alimpay:$(VERSION) .
	@echo "Docker image built: alimpay:$(VERSION)"

# Docker运行
docker-run:
	@echo "Running Docker container..."
	docker run -d -p 8080:8080 --name alimpay alimpay:$(VERSION)

# 生成API文档
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/alimpay/main.go
	@echo "API documentation generated"

