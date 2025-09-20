.PHONY: build test clean run-node run-wallet help

# Переменные
BINARY_NAME=mirochain
WALLET_BINARY=mirochain-wallet
BUILD_DIR=build
GO_VERSION=1.25

# Цели по умолчанию
all: build

# Сборка всех бинарных файлов
build: build-node build-wallet

# Сборка узла блокчейна
build-node:
	@echo "Building blockchain node..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/node/main.go
	@echo "Node binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Сборка CLI кошелька
build-wallet:
	@echo "Building wallet CLI..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(WALLET_BINARY) cmd/wallet/main.go
	@echo "Wallet binary built: $(BUILD_DIR)/$(WALLET_BINARY)"

# Запуск тестов
test:
	@echo "Running tests..."
	@go test -v ./tests/...

# Запуск узла блокчейна
run-node:
	@echo "Starting blockchain node..."
	@go run cmd/node/main.go

# Запуск CLI кошелька
run-wallet:
	@echo "Starting wallet CLI..."
	@go run cmd/wallet/main.go

# Создание нового кошелька
create-wallet:
	@echo "Creating new wallet..."
	@go run cmd/wallet/main.go -create

# Список кошельков
list-wallets:
	@echo "Listing wallets..."
	@go run cmd/wallet/main.go -list

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf data/
	@rm -rf wallet_data/
	@rm -rf logs/
	@echo "Clean completed"

# Установка зависимостей
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Проверка кода
lint:
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...

# Запуск с майнингом
mine:
	@echo "Starting node with mining enabled..."
	@go run cmd/node/main.go -mining=true -difficulty=4

# Запуск без майнинга
no-mine:
	@echo "Starting node without mining..."
	@go run cmd/node/main.go -mining=false

# Показать справку
help:
	@echo "MiroChain Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build all binaries"
	@echo "  build-node     - Build blockchain node"
	@echo "  build-wallet   - Build wallet CLI"
	@echo "  test           - Run tests"
	@echo "  run-node       - Run blockchain node"
	@echo "  run-wallet     - Run wallet CLI"
	@echo "  create-wallet  - Create new wallet"
	@echo "  list-wallets   - List all wallets"
	@echo "  mine           - Run node with mining"
	@echo "  no-mine        - Run node without mining"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  lint           - Run linter"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make mine"
	@echo "  make create-wallet"
