# Makefile for Shift Planner Project

# Variables
BINARY_NAME=server
BACKEND_DIR=backend
FRONTEND_DIR=frontend
CMD_DIR=$(BACKEND_DIR)/cmd/server
BUILD_DIR=build
PORT=8080
FRONTEND_PORT=3000
PID_FILE=.pids

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    MKDIR := if not exist $(BUILD_DIR) mkdir $(BUILD_DIR)
    RMDIR := if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
    MKDIR_DATA := if not exist data mkdir data
    BINARY_EXT := .exe
    KILL_BACKEND := powershell -Command "Get-NetTCPConnection -LocalPort $(PORT) -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id $$_.OwningProcess -Force -ErrorAction SilentlyContinue }" 2>nul & taskkill /F /IM go.exe 2>nul || echo Backend stopped
    KILL_FRONTEND := powershell -Command "Get-NetTCPConnection -LocalPort $(FRONTEND_PORT) -ErrorAction SilentlyContinue | ForEach-Object { Stop-Process -Id $$_.OwningProcess -Force -ErrorAction SilentlyContinue }" 2>nul & taskkill /F /IM node.exe 2>nul || echo Frontend stopped
    CHECK_BACKEND := tasklist /FI "IMAGENAMmakjeE eq go.exe" 2>nul | find /I "go.exe" >nul
    CHECK_FRONTEND := tasklist /FI "IMAGENAME eq node.exe" 2>nul | find /I "node.exe" >nul
else
    DETECTED_OS := Unix
    MKDIR := mkdir -p $(BUILD_DIR)
    RMDIR := rm -rf $(BUILD_DIR)
    MKDIR_DATA := mkdir -p data
    BINARY_EXT :=
    KILL_BACKEND := pkill -f "go run.*cmd/server" || true
    KILL_FRONTEND := pkill -f "vite" || true
    CHECK_BACKEND := pgrep -f "go run.*cmd/server" > /dev/null
    CHECK_FRONTEND := pgrep -f "vite" > /dev/null
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

.PHONY: all build clean test deps run install help dev fmt lint init setup up down status

# Default target - run both backend and frontend
.DEFAULT_GOAL := up

# Help target
help:
	@echo "Shift Planner - Makefile Commands:"
	@echo ""
	@echo "  make              - Start both backend and frontend (default)"
	@echo "  make up           - Start both backend and frontend"
	@echo "  make down         - Stop both backend and frontend"
	@echo "  make status       - Check status of backend and frontend"
	@echo "  make build        - Build the backend server"
	@echo "  make run          - Run only the backend server"
	@echo "  make frontend     - Run only the frontend"
	@echo "  make test         - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make deps         - Download and tidy dependencies"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install      - Install dependencies and build"
	@echo "  make fmt          - Format Go code"
	@echo "  make lint         - Lint Go code"
	@echo "  make init         - Initialize project"
	@echo "  make setup        - Full setup: init, deps, build"
	@echo ""

# Start both backend and frontend
up:
	@echo "Starting Shift Planner..."
	@echo "Backend: http://localhost:$(PORT)"
	@echo "Frontend: http://localhost:$(FRONTEND_PORT)"
	@echo ""
ifeq ($(OS),Windows_NT)
	@cd $(BACKEND_DIR) && start cmd /k "$(GOCMD) run ./cmd/server"
	@timeout /t 3 /nobreak >nul
	@cd $(FRONTEND_DIR) && start cmd /k "npm run dev"
	@echo Both servers started in separate windows
	@echo Use 'make down' to stop both servers
else
	@cd $(BACKEND_DIR) && $(GOCMD) run ./cmd/server > ../backend.log 2>&1 & echo $$! > ../backend.pid
	@sleep 2
	@cd $(FRONTEND_DIR) && npm run dev > ../frontend.log 2>&1 & echo $$! > ../frontend.pid
	@echo "Both servers started"
	@echo "Backend PID: $$(cat backend.pid)"
	@echo "Frontend PID: $$(cat frontend.pid)"
	@echo "Use 'make down' to stop both servers"
endif

# Stop both backend and frontend
down:
	@echo "Stopping Shift Planner..."
ifeq ($(OS),Windows_NT)
	@echo "Stopping backend..."
	@$(KILL_BACKEND)
	@echo "Stopping frontend..."
	@$(KILL_FRONTEND)
	@echo "Both servers stopped"
else
	@if [ -f backend.pid ]; then kill $$(cat backend.pid) 2>/dev/null || true; rm -f backend.pid; fi
	@if [ -f frontend.pid ]; then kill $$(cat frontend.pid) 2>/dev/null || true; rm -f frontend.pid; fi
	@$(KILL_BACKEND)
	@$(KILL_FRONTEND)
	@echo "Both servers stopped"
endif

# Check status
status:
	@echo Checking server status...
ifeq ($(OS),Windows_NT)
	@echo Backend: 
	@$(CHECK_BACKEND) && echo Running || echo Stopped
	@echo Frontend: 
	@$(CHECK_FRONTEND) && echo Running || echo Stopped
else
	@echo -n "Backend: "
	@$(CHECK_BACKEND) && echo "Running" || echo "Stopped"
	@echo -n "Frontend: "
	@$(CHECK_FRONTEND) && echo "Running" || echo "Stopped"
endif

# Build the backend server
build:
	@echo "Building backend server..."
	@$(MKDIR)
	@cd $(BACKEND_DIR) && $(GOBUILD) $(LDFLAGS) -o ../$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

# Run only the backend server (API only)
run:
	@echo "Starting backend API server on port $(PORT)..."
	@echo "API endpoints available at: http://localhost:$(PORT)/api"
	@echo "Note: Frontend should run separately on port $(FRONTEND_PORT)"
	@cd $(BACKEND_DIR) && $(GOCMD) run ./cmd/server

# Run only the frontend
frontend:
	@echo "Starting frontend on port $(FRONTEND_PORT)..."
	@echo "Open http://localhost:$(FRONTEND_PORT) in your browser"
	@cd $(FRONTEND_DIR) && npm run dev

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
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && npm install
	@echo "Dependencies updated"

# Install dependencies and build
install: deps build

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@$(RMDIR)
	@cd $(BACKEND_DIR) && $(GOCLEAN) -cache
	@rm -f backend.pid frontend.pid backend.log frontend.log 2>/dev/null || true
	@echo "Clean complete"

# Development mode (requires air for hot reload)
dev: up

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
	@echo "Run 'make up' to start both servers"
