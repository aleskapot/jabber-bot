.PHONY: build run test clean deps test-integration build-all

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Windows_NT)
	BUILD_EXT := .exe
	RUN_CMD := bin/jabber-bot.exe
	TEST_SCRIPT := scripts/run-tests.bat
	INTEGRATION_SCRIPT := scripts/run-integration-tests.bat
	BUILD_SCRIPT := scripts/build.bat
else
	BUILD_EXT := 
	RUN_CMD := bin/jabber-bot
	TEST_SCRIPT := ./scripts/run-tests.sh
	INTEGRATION_SCRIPT := ./scripts/run-integration-tests.sh
	BUILD_SCRIPT := make build-go
endif

# Базовые команды
build: build-go

build-go:
	@echo "🏗️  Building Jabber Bot..."
ifeq ($(UNAME_S),Windows_NT)
	powershell -Command "if (!(Test-Path bin)) { mkdir bin }"
else
	mkdir -p bin
endif
	go build -o bin/jabber-bot$(BUILD_EXT) ./cmd/server
	@echo "✅ Build completed: bin/jabber-bot$(BUILD_EXT)"

build-all:
ifeq ($(UNAME_S),Windows_NT)
	@echo "🏗️  Building for all platforms..."
	$(BUILD_SCRIPT) --all
else
	@echo "🏗️  Building for all platforms..."
	@echo "🐧 Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o bin/jabber-bot-linux-amd64 ./cmd/server
	@echo "🐧 Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o bin/jabber-bot-linux-arm64 ./cmd/server
	@echo "🍎 macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o bin/jabber-bot-darwin-amd64 ./cmd/server
	@echo "🍎 macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o bin/jabber-bot-darwin-arm64 ./cmd/server
	@echo "🪟 Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o bin/jabber-bot-windows-amd64.exe ./cmd/server
	@echo "✅ All builds completed in bin/"
endif

run: build
ifeq ($(UNAME_S),Windows_NT)
	$(RUN_CMD) -config configs/config.yaml
else
	$(RUN_CMD) -config configs/config.yaml
endif

test:
	@echo "🧪 Running unit tests..."
	go test ./...

test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-integration:
	@echo "🌐 Running integration tests..."
	INTEGRATION_TESTS=1 go test -tags=integration ./test/integration/...

test-all:
	@echo "🧪 Running all tests..."
ifeq ($(UNAME_S),Windows_NT)
	$(TEST_SCRIPT)
else
	$(TEST_SCRIPT)
endif

# Очистка
clean:
	@echo "🧹 Cleaning up..."
ifeq ($(UNAME_S),Windows_NT)
	if exist bin rmdir /s /q bin
	if exist coverage.out del coverage.out
	if exist coverage.html del coverage.html
else
	rm -rf bin/
	rm -f coverage.out coverage.html
endif

# Зависимости
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Форматирование
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Линтинг
lint:
	@echo "🔍 Running linter..."
	golangci-lint run

# Генерация зависимостей
generate:
	@echo "🔧 Generating code..."
	go generate ./...

# Docker
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t jabber-bot .

docker-run:
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d

docker-stop:
	@echo "🐳 Stopping Docker containers..."
	docker-compose down

# Установка инструментов
install-tools:
	@echo "🛠️  Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Fast test commands for development
quick-test:
	@echo "⚡ Quick unit tests (no coverage)..."
	go test -short ./...

quick-integration:
	@echo "⚡ Quick integration tests..."
	INTEGRATION_TESTS=1 go test -short -tags=integration ./test/integration/...