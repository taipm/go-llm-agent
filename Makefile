.PHONY: help build test clean examples run-simple run-tools run-conv

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	go test -v ./pkg/...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.txt -covermode=atomic ./pkg/...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build all examples
	@echo "Building examples..."
	@go build -o bin/simple_chat ./examples/simple_chat
	@go build -o bin/tool_usage ./examples/tool_usage
	@go build -o bin/conversation ./examples/conversation
	@echo "Build complete! Binaries in ./bin/"

clean: ## Clean build artifacts
	@rm -rf bin/
	@rm -f coverage.txt coverage.html
	@echo "Clean complete!"

run-simple: ## Run simple chat example
	go run ./examples/simple_chat/main.go

run-tools: ## Run tool usage example
	go run ./examples/tool_usage/main.go

run-conv: ## Run conversation example
	go run ./examples/conversation/main.go

run-stream: ## Run streaming example
	go run ./examples/streaming/main.go

run-stream-adv: ## Run advanced streaming example
	go run ./examples/streaming_advanced/main.go

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run linters

tidy: ## Tidy dependencies
	go mod tidy

check: lint test ## Run all checks (lint + test)

all: clean tidy check build ## Run all tasks
