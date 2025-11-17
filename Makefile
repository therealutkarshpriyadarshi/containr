.PHONY: build clean install test test-unit test-integration test-coverage test-all run-example fmt lint lint-ci deps help bench profile-cpu profile-mem profile-trace profile-all release

# Build variables
BINARY_NAME=containr
BUILD_DIR=bin
GO=go
VERSION?=1.0.0
GIT_COMMIT?=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOFLAGS=-ldflags="-s -w -X github.com/therealutkarshpriyadarshi/containr/pkg/version.Version=$(VERSION) -X github.com/therealutkarshpriyadarshi/containr/pkg/version.GitCommit=$(GIT_COMMIT) -X github.com/therealutkarshpriyadarshi/containr/pkg/version.BuildDate=$(BUILD_DATE)"
COVERAGE_DIR=coverage
PROFILE_DIR=profiles

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

# Phase 4: Benchmarking and Profiling

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GO) test -bench=. -benchmem -benchtime=1s ./pkg/...

# Run benchmarks with CPU profile
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	@mkdir -p $(PROFILE_DIR)
	@$(GO) test -bench=. -benchmem -cpuprofile=$(PROFILE_DIR)/bench-cpu.prof ./pkg/...
	@echo "CPU profile saved to $(PROFILE_DIR)/bench-cpu.prof"

# Run benchmarks with memory profile
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	@mkdir -p $(PROFILE_DIR)
	@$(GO) test -bench=. -benchmem -memprofile=$(PROFILE_DIR)/bench-mem.prof ./pkg/...
	@echo "Memory profile saved to $(PROFILE_DIR)/bench-mem.prof"

# Generate CPU profile
profile-cpu:
	@echo "Generating CPU profile..."
	@mkdir -p $(PROFILE_DIR)
	@echo "Run your containr commands, then press Ctrl+C"
	@$(GO) test -cpuprofile=$(PROFILE_DIR)/cpu.prof -run=XXX -bench=. ./pkg/benchmark
	@echo "CPU profile saved to $(PROFILE_DIR)/cpu.prof"
	@echo "Analyze with: go tool pprof $(PROFILE_DIR)/cpu.prof"

# Generate memory profile
profile-mem:
	@echo "Generating memory profile..."
	@mkdir -p $(PROFILE_DIR)
	@$(GO) test -memprofile=$(PROFILE_DIR)/mem.prof -run=XXX -bench=. ./pkg/benchmark
	@echo "Memory profile saved to $(PROFILE_DIR)/mem.prof"
	@echo "Analyze with: go tool pprof $(PROFILE_DIR)/mem.prof"

# Generate execution trace
profile-trace:
	@echo "Generating execution trace..."
	@mkdir -p $(PROFILE_DIR)
	@$(GO) test -trace=$(PROFILE_DIR)/trace.out -run=TestProfiler ./pkg/profiler
	@echo "Trace saved to $(PROFILE_DIR)/trace.out"
	@echo "Analyze with: go tool trace $(PROFILE_DIR)/trace.out"

# Generate all profiles
profile-all: profile-cpu profile-mem profile-trace
	@echo "All profiles generated in $(PROFILE_DIR)/"

# View CPU profile in browser
profile-view-cpu:
	@echo "Opening CPU profile in browser..."
	@go tool pprof -http=:8080 $(PROFILE_DIR)/cpu.prof

# View memory profile in browser
profile-view-mem:
	@echo "Opening memory profile in browser..."
	@go tool pprof -http=:8080 $(PROFILE_DIR)/mem.prof

# Clean profile data
profile-clean:
	@echo "Cleaning profile data..."
	@rm -rf $(PROFILE_DIR)
	@echo "Profile data cleaned"

# Phase 4: Release Management

# Build release binaries for all platforms
release:
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release
	@echo "Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 ./cmd/containr
	@echo "Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 ./cmd/containr
	@echo "Building for linux/arm..."
	@GOOS=linux GOARCH=arm $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm ./cmd/containr
	@echo "Release binaries built in $(BUILD_DIR)/release/"

# Create release archive
release-archive: release
	@echo "Creating release archives..."
	@cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	@cd $(BUILD_DIR)/release && tar -czf $(BINARY_NAME)-$(VERSION)-linux-arm.tar.gz $(BINARY_NAME)-linux-arm
	@cd $(BUILD_DIR)/release && sha256sum *.tar.gz > checksums.txt
	@echo "Release archives created with checksums"

# Verify release
release-verify:
	@echo "Verifying release..."
	@./$(BUILD_DIR)/$(BINARY_NAME) version
	@$(GO) version
	@echo "Release verification complete"

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build            - Build the containr binary"
	@echo "  build-example    - Build the example program"
	@echo "  install          - Install containr to /usr/local/bin"
	@echo "  clean            - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests (requires root)"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  test-all         - Run all tests with coverage"
	@echo ""
	@echo "Benchmarking (Phase 4):"
	@echo "  bench            - Run benchmarks"
	@echo "  bench-cpu        - Run benchmarks with CPU profiling"
	@echo "  bench-mem        - Run benchmarks with memory profiling"
	@echo ""
	@echo "Profiling (Phase 4):"
	@echo "  profile-cpu      - Generate CPU profile"
	@echo "  profile-mem      - Generate memory profile"
	@echo "  profile-trace    - Generate execution trace"
	@echo "  profile-all      - Generate all profiles"
	@echo "  profile-view-cpu - View CPU profile in browser"
	@echo "  profile-view-mem - View memory profile in browser"
	@echo "  profile-clean    - Clean profile data"
	@echo ""
	@echo "Release (Phase 4):"
	@echo "  release          - Build release binaries for all platforms"
	@echo "  release-archive  - Create release archives with checksums"
	@echo "  release-verify   - Verify release build"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Check code formatting"
	@echo "  lint             - Lint code"
	@echo "  lint-ci          - Lint code (CI mode, strict)"
	@echo "  staticcheck      - Run static analysis"
	@echo "  security         - Run security scan"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps             - Download dependencies"
	@echo "  deps-verify      - Verify dependencies"
	@echo "  deps-update      - Update dependencies"
	@echo "  install-tools    - Install development tools"
	@echo ""
	@echo "Workflows:"
	@echo "  pre-commit       - Run pre-commit checks"
	@echo "  ci               - Run full CI checks"
	@echo "  run-example      - Run the example (requires root)"
	@echo ""
	@echo "  help             - Show this help message"
