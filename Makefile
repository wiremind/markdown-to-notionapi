.PHONY: help build test lint clean install deps security fmt vet check coverage run-local container-build container-test

# Default Go version
GO_VERSION ?= 1.25
BINARY_NAME = md2notion
MAIN_PATH = ./cmd/notion-md
DIST_DIR = dist

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
NC = \033[0m # No Color

help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

deps: ## Download and verify dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	go mod download
	go mod verify
	go mod tidy

build: deps ## Build the binary
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	mkdir -p $(DIST_DIR)
	go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Binary built: $(DIST_DIR)/$(BINARY_NAME)$(NC)"

build-all: deps ## Build binaries for all platforms
	@echo "$(YELLOW)Building for all platforms...$(NC)"
	mkdir -p $(DIST_DIR)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build -ldflags='-s -w' -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "$(GREEN)All binaries built in $(DIST_DIR)/$(NC)"

install: build ## Install the binary to GOPATH/bin
	@echo "$(YELLOW)Installing $(BINARY_NAME)...$(NC)"
	go install $(MAIN_PATH)
	@echo "$(GREEN)$(BINARY_NAME) installed to $(shell go env GOPATH)/bin$(NC)"

test: deps ## Run tests
	@echo "$(YELLOW)Running tests...$(NC)"
	go test -v -race ./...

coverage: deps ## Run tests with coverage
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)goimports not found, skipping import formatting$(NC)"; \
	fi

vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	go vet ./...

lint: ## Run golangci-lint
	@echo "$(YELLOW)Running golangci-lint...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(RED)golangci-lint not installed. Run: make install-lint$(NC)"; \
	fi

install-lint: ## Install golangci-lint
	@echo "$(YELLOW)Installing golangci-lint...$(NC)"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

security: ## Run security checks
	@echo "$(YELLOW)Running security checks...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(RED)gosec not installed. Run: make install-security$(NC)"; \
	fi
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(RED)govulncheck not installed. Run: make install-security$(NC)"; \
	fi

install-security: ## Install security tools
	@echo "$(YELLOW)Installing security tools...$(NC)"
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

check: fmt vet lint test security ## Run all checks (format, vet, lint, test, security)

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -testcache
	@echo "$(GREEN)Clean complete$(NC)"

# Container targets
container-build: ## Build container image
	@echo "$(YELLOW)Building container image...$(NC)"
	nerdctl build -f Containerfile -t $(BINARY_NAME):latest .
	@echo "$(GREEN)Container image built: $(BINARY_NAME):latest$(NC)"

container-test: container-build ## Test container image
	@echo "$(YELLOW)Testing container image...$(NC)"
	@if [ -z "$$NOTION_TOKEN" ]; then \
		echo "$(RED)Error: NOTION_TOKEN environment variable is required$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$$TEST_PAGE_ID" ]; then \
		echo "$(RED)Error: TEST_PAGE_ID environment variable is required$(NC)"; \
		echo "$(YELLOW)Set it to a Notion page ID for testing$(NC)"; \
		exit 1; \
	fi
	echo "Testing container with dry-run..."
	nerdctl run --rm -e NOTION_TOKEN=$$NOTION_TOKEN -v $$(pwd):/data $(BINARY_NAME):latest -page-id=$$TEST_PAGE_ID -dry-run /data/README.md
	@echo "$(GREEN)Container test passed!$(NC)"

# Development helpers
run-local: build ## Run locally with test data
	@echo "$(YELLOW)Running locally...$(NC)"
	@if [ -z "$$NOTION_TOKEN" ]; then \
		echo "$(RED)Error: NOTION_TOKEN environment variable is required$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$$TEST_PAGE_ID" ]; then \
		echo "$(RED)Error: TEST_PAGE_ID environment variable is required$(NC)"; \
		exit 1; \
	fi
	$(DIST_DIR)/$(BINARY_NAME) -page-id=$$TEST_PAGE_ID -dry-run README.md

dev-setup: deps install-lint install-security ## Set up development environment
	@echo "$(GREEN)Development environment setup complete!$(NC)"
	@echo "$(YELLOW)Available commands:$(NC)"
	@make help
