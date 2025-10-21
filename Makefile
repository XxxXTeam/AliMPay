.PHONY: build run clean test init help

# 变量定义
BINARY_NAME=alimpay
CONFIG_PATH=./configs/config.yaml

# 默认目标
help:
	@echo "AliMPay Golang Version - Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - 编译项目"
	@echo "  make run      - 运行项目"
	@echo "  make init     - 初始化数据库"
	@echo "  make clean    - 清理编译文件"
	@echo "  make test     - 运行测试"
	@echo "  make install  - 安装依赖"
	@echo ""

# 安装依赖
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 编译项目
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) cmd/alimpay/main.go
	@echo "Build complete: $(BINARY_NAME)"

# 运行项目
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME) -config=$(CONFIG_PATH)

# 初始化数据库
init:
	@echo "Initializing database..."
	go run scripts/init_db.go -config=$(CONFIG_PATH)

# 清理编译文件
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf data/
	rm -rf logs/
	@echo "Clean complete"

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Linting code..."
	golangci-lint run

# 开发模式运行
dev:
	@echo "Running in development mode..."
	GIN_MODE=debug go run cmd/alimpay/main.go -config=$(CONFIG_PATH)

