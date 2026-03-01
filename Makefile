.PHONY: build clean test test-unit test-integration test-e2e test-coverage lint fmt vet

# 构建配置
BINARY_NAME=chopsticks
BINARY_DIR=bin
CMD_PATH=./cmd/cli

# Go 配置
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 默认目标
all: build

# 构建二进制文件
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME).exe $(CMD_PATH)
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME).exe"

# 清理构建产物
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# 运行所有测试
test: test-unit test-integration
	@echo "All tests passed"

# 运行单元测试
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -race -count=1 ./core/... ./engine/... ./pkg/...

# 运行集成测试
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -race -count=1 -tags=integration ./test/integration/...

# 运行 E2E 测试
test-e2e:
	@echo "Running E2E tests..."
	$(GOTEST) -v -count=1 -tags=e2e ./test/e2e/...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 运行特定集成测试
test-integration-specific:
	@echo "Running specific integration test: $(TEST)"
	$(GOTEST) -v -tags=integration ./test/integration/... -run $(TEST)

# 代码格式化
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 代码检查
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# 使用 golangci-lint 进行代码检查
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet instead..."; \
		$(GOCMD) vet ./...; \
	fi

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 运行应用
run: build
	$(BINARY_DIR)/$(BINARY_NAME).exe

# 安装到系统
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_DIR)/$(BINARY_NAME).exe $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME).exe"

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run all tests (unit + integration)"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e       - Run E2E tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linter"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  run            - Build and run the application"
	@echo "  help           - Show this help message"
