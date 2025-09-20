# MiroChain Makefile

# Variables
APP_NAME = mirochain
VERSION = $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION = $(shell go version | cut -d' ' -f3)
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# Docker variables
DOCKER_IMAGE = mirochain
DOCKER_TAG = latest
DOCKER_REGISTRY = 

# Kubernetes variables
K8S_NAMESPACE = mirochain

# Default target
.PHONY: all
all: build

# Build targets
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) cmd/node/main.go

.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 cmd/node/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 cmd/node/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe cmd/node/main.go

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Lint targets
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	@echo "Fixing linting issues..."
	golangci-lint run --fix

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf data/
	rm -rf profiles/

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REGISTRY)$(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 8081:8081 -p 8082:8082 -p 8083:8083 -p 8084:8084 -p 8085:8085 -p 8086:8086 -p 8087:8087 -p 8088:8088 $(DOCKER_REGISTRY)$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing docker-compose logs..."
	docker-compose logs -f

# Kubernetes targets
.PHONY: k8s-apply
k8s-apply:
	@echo "Applying Kubernetes manifests..."
	kubectl apply -f k8s/

.PHONY: k8s-delete
k8s-delete:
	@echo "Deleting Kubernetes resources..."
	kubectl delete -f k8s/

.PHONY: k8s-status
k8s-status:
	@echo "Checking Kubernetes status..."
	kubectl get pods -n $(K8S_NAMESPACE)
	kubectl get services -n $(K8S_NAMESPACE)
	kubectl get ingress -n $(K8S_NAMESPACE)

# Development targets
.PHONY: dev
dev:
	@echo "Starting development server..."
	go run cmd/node/main.go -port=8080 -mining=false

.PHONY: dev-mining
dev-mining:
	@echo "Starting development server with mining..."
	go run cmd/node/main.go -port=8080 -mining=true

.PHONY: dev-multi
dev-multi:
	@echo "Starting multiple development nodes..."
	@echo "Starting node 1..."
	go run cmd/node/main.go -port=8080 -mining=false &
	@echo "Starting node 2..."
	go run cmd/node/main.go -port=9080 -mining=true &
	@echo "Starting node 3..."
	go run cmd/node/main.go -port=10080 -mining=false &
	@echo "All nodes started. Press Ctrl+C to stop all."

# Demo targets
.PHONY: demo
demo:
	@echo "Running demos..."
	@echo "Running GraphQL demo..."
	go run -tags graphql_demo examples/graphql_demo.go

.PHONY: demo-contracts
demo-contracts:
	@echo "Running smart contracts demo..."
	go run -tags contract_demo examples/contract_demo.go

.PHONY: demo-tokens
demo-tokens:
	@echo "Running tokens demo..."
	go run -tags token_demo examples/token_demo.go

.PHONY: demo-nfts
demo-nfts:
	@echo "Running NFTs demo..."
	go run -tags nft_demo examples/nft_demo.go

.PHONY: demo-sidechains
demo-sidechains:
	@echo "Running sidechains demo..."
	go run -tags sidechain_demo examples/sidechain_demo.go

.PHONY: demo-statechannels
demo-statechannels:
	@echo "Running state channels demo..."
	go run -tags statechannel_demo examples/statechannel_demo.go

# Documentation targets
.PHONY: docs
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/node/main.go -o docs/swagger

# Security targets
.PHONY: security-scan
security-scan:
	@echo "Running security scan..."
	gosec ./...

.PHONY: dependency-check
dependency-check:
	@echo "Checking dependencies..."
	go list -json -m all | nancy sleuth

# Performance targets
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: profile
profile:
	@echo "Running with profiling..."
	go run cmd/node/main.go -profiling=true

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-all      - Build for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-race      - Run tests with race detection"
	@echo "  lint           - Run linter"
	@echo "  lint-fix       - Fix linting issues"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-compose-up   - Start services with docker-compose"
	@echo "  docker-compose-down  - Stop services with docker-compose"
	@echo "  k8s-apply      - Apply Kubernetes manifests"
	@echo "  k8s-delete     - Delete Kubernetes resources"
	@echo "  k8s-status     - Check Kubernetes status"
	@echo "  dev            - Start development server"
	@echo "  dev-mining     - Start development server with mining"
	@echo "  dev-multi      - Start multiple development nodes"
	@echo "  demo           - Run all demos"
	@echo "  demo-contracts - Run smart contracts demo"
	@echo "  demo-tokens    - Run tokens demo"
	@echo "  demo-nfts      - Run NFTs demo"
	@echo "  demo-sidechains - Run sidechains demo"
	@echo "  demo-statechannels - Run state channels demo"
	@echo "  docs           - Generate documentation"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  security-scan  - Run security scan"
	@echo "  dependency-check - Check dependencies"
	@echo "  bench          - Run benchmarks"
	@echo "  profile        - Run with profiling"
	@echo "  help           - Show this help message"