# Makefile for Shift Planner Project

# Variables
BINARY_NAME=server
BACKEND_DIR=backend
CMD_DIR=$(BACKEND_DIR)/cmd/server
BUILD_DIR=build
PORT=8080

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    MKDIR := if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
    RMDIR := if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
    MKDIR_DATA := if not exist data mkdir data
    BINARY_EXT := .exe
else
    DETECTED_OS := Unix
    MKDIR := mkdir -p $(BUILD_DIR)
    RMDIR := rm -rf $(BUILD_DIR)
    MKDIR_DATA := mkdir -p data
    BINARY_EXT :=
endif

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test deps run install help dev fmt lint init setup

# Default target
all: clean deps build

# Help target
help:
	@echo "Shift Planner - Makefile Commands:"
	@echo ""
	@echo "  make build          - Build the backend server"
	@echo "  make run            - Run the backend server"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install dependencies and build"
	@echo "  make dev            - Run in development mode (with auto-reload)"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Lint Go code (requires golangci-lint)"
	@echo "  make init           - Initialize project (create data directory)"
	@echo "  make setup          - Full setup: init, deps, build"
	@echo ""

# Build the backend server
build:
	@echo "Building backend server..."
	@$(MKDIR)
	@cd $(BACKEND_DIR) && $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

# Run the backend server
run:
	@echo "Starting server on port $(PORT)..."
	@echo "Open http://localhost:$(PORT) in your browser"
	@cd $(BACKEND_DIR) && $(GOCMD) run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	@cd $(BACKEND_DIR) && $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@cd $(BACKEND_DIR) && $(GOTEST) -v -coverprofile=coverage.out ./...
	@cd $(BACKEND_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: $(BACKEND_DIR)/coverage.html"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@cd $(BACKEND_DIR) && $(GOMOD) download
	@cd $(BACKEND_DIR) && $(GOMOD) tidy
	@echo "Dependencies updated"

# Install dependencies and build
install: deps build

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@$(RMDIR)
	@cd $(BACKEND_DIR) && $(GOCLEAN) -cache
	@echo "Clean complete"

# Development mode (requires air for hot reload)
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null 2>&1; then \
		cd $(BACKEND_DIR) && air; \
	else \
		echo "Air not found. Installing air..."; \
		$(GOINSTALL) github.com/cosmtrek/air@latest; \
		cd $(BACKEND_DIR) && air; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@cd $(BACKEND_DIR) && $(GOCMD) fmt ./...
	@echo "Formatting complete"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		cd $(BACKEND_DIR) && golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
		echo "Or use: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Create data directory if it doesn't exist
init:
	@echo "Initializing project..."
	@$(MKDIR_DATA)
	@echo "Project initialized"

# Full setup: init, deps, build
setup: init deps build
	@echo "Setup complete!"
	@echo "Run 'make run' to start the server"
