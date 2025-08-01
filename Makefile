.PHONY: build run test clean deps migrate

# 构建变量
BINARY_NAME=keydesk
BUILD_DIR=build
MAIN_PATH=cmd/server/main.go

# 默认目标
all: deps build

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建应用
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行应用
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式运行
dev:
	@echo "Running in development mode..."
	go run $(MAIN_PATH)

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行示例
example:
	@echo "Running example..."
	go run examples/basic_usage.go

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	# 这里可以添加数据库迁移命令

# 清理构建文件
clean:
	@echo "Cleaning build files..."
	rm -rf $(BUILD_DIR)
	go clean

# 安装到系统
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# 生成API文档
docs:
	@echo "Generating API documentation..."
	# 这里可以添加API文档生成命令

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 创建发布版本
release: clean build
	@echo "Creating release..."
	tar -czf $(BINARY_NAME)-$(shell git describe --tags --always).tar.gz $(BUILD_DIR)/

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  run      - Build and run the application"
	@echo "  dev      - Run in development mode"
	@echo "  test     - Run tests"
	@echo "  example  - Run example code"
	@echo "  clean    - Clean build files"
	@echo "  install  - Install to system"
	@echo "  fmt      - Format code"
	@echo "  lint     - Run linter"
	@echo "  release  - Create release package" 