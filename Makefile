.PHONY: build clean install test run-example

# Build variables
BINARY_NAME=containr
BUILD_DIR=bin
GO=go
GOFLAGS=-ldflags="-s -w"

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
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v ./...

# Run example (requires root)
run-example: build-example
	@echo "Running example (requires root)..."
	@sudo $(BUILD_DIR)/example

# Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golint ./...

# Get dependencies
deps:
	@echo "Getting dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the containr binary"
	@echo "  build-example - Build the example program"
	@echo "  install      - Install containr to /usr/local/bin"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  run-example  - Run the example (requires root)"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  deps         - Download dependencies"
	@echo "  help         - Show this help message"
