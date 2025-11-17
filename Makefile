.PHONY: build clean install test test-unit test-integration test-coverage test-all run-example fmt lint lint-ci deps help

# Build variables
BINARY_NAME=containr
BUILD_DIR=bin
GO=go
GOFLAGS=-ldflags="-s -w"
COVERAGE_DIR=coverage

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/containr
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build example
build-example:
	@echo "Building example..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/example ./examples/simple.go
	@echo "Example built: $(BUILD_DIR)/example"

# Install to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	@$(GO) clean
	@echo "Clean complete"

# Run all tests
test: test-unit test-integration
	@echo "All tests completed"

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@$(GO) test -v -race ./pkg/...

# Run integration tests (requires root)
test-integration:
	@echo "Running integration tests (requires root)..."
	@sudo $(GO) test -v ./test/... || echo "Integration tests may require root privileges"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./pkg/...
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -n 1
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

# Run all tests with coverage
test-all: test-coverage test-integration
	@echo "All tests with coverage completed"

# Run example (requires root)
run-example: build-example
	@echo "Running example (requires root)..."
	@sudo $(BUILD_DIR)/example

# Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@gofmt -s -w .

# Check code formatting
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)" || (echo "Please run 'make fmt' to format code" && exit 1)

# Lint code
lint:
	@echo "Linting code..."
	@$(GO) vet ./...
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

# Lint for CI (fail on errors)
lint-ci:
	@echo "Linting code (CI mode)..."
	@$(GO) vet ./...
	@golangci-lint run ./...

# Static analysis
staticcheck:
	@echo "Running static analysis..."
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, run: go install honnef.co/go/tools/cmd/staticcheck@latest"

# Security scan
security:
	@echo "Running security scan..."
	@which gosec > /dev/null 2>&1 && gosec ./... || echo "gosec not installed, run: go install github.com/securego/gosec/v2/cmd/gosec@latest"

# Get dependencies
deps:
	@echo "Getting dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	@$(GO) mod verify

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Quick check before commit
pre-commit: fmt-check lint test-unit
	@echo "Pre-commit checks passed!"

# Full CI check
ci: deps lint-ci test-coverage test-integration staticcheck security
	@echo "CI checks completed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the containr binary"
	@echo "  build-example    - Build the example program"
	@echo "  install          - Install containr to /usr/local/bin"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests (requires root)"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-all         - Run all tests with coverage"
	@echo "  run-example      - Run the example (requires root)"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Check code formatting"
	@echo "  lint             - Lint code"
	@echo "  lint-ci          - Lint code (CI mode, strict)"
	@echo "  staticcheck      - Run static analysis"
	@echo "  security         - Run security scan"
	@echo "  deps             - Download dependencies"
	@echo "  deps-verify      - Verify dependencies"
	@echo "  deps-update      - Update dependencies"
	@echo "  install-tools    - Install development tools"
	@echo "  pre-commit       - Run pre-commit checks"
	@echo "  ci               - Run full CI checks"
	@echo "  help             - Show this help message"
