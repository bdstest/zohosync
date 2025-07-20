# ZohoSync Makefile

APP_NAME := zohosync
CLI_NAME := $(APP_NAME)-cli
DAEMON_NAME := $(APP_NAME)-daemon

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.commit=$(COMMIT)

.PHONY: all build clean test lint install security security-scan security-install security-quick

all: build

build: build-gui build-cli build-daemon

build-gui:
	@echo "Building GUI application..."
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) ./cmd/gui

build-cli:
	@echo "Building CLI application..."
	go build -ldflags "$(LDFLAGS)" -o $(CLI_NAME) ./cmd/cli

build-daemon:
	@echo "Building daemon..."
	go build -ldflags "$(LDFLAGS)" -o $(DAEMON_NAME) ./cmd/daemon

clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME) $(CLI_NAME) $(DAEMON_NAME)
	rm -rf dist/ build/

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

lint:
	@echo "Running linter..."
	golangci-lint run

install: build
	@echo "Installing..."
	sudo cp $(APP_NAME) /usr/local/bin/
	sudo cp $(CLI_NAME) /usr/local/bin/
	sudo cp $(DAEMON_NAME) /usr/local/bin/

# Security Testing Targets
security: security-install security-scan

security-install:
	@echo "Installing security tools..."
	@command -v gosec >/dev/null 2>&1 || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@command -v staticcheck >/dev/null 2>&1 || go install honnef.co/go/tools/cmd/staticcheck@latest
	@command -v nancy >/dev/null 2>&1 || go install github.com/sonatypecommunity/nancy@latest
	@command -v golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2

security-scan:
	@echo "Running comprehensive security analysis..."
	@./security/security-scan.sh

security-quick:
	@echo "Running quick security checks..."
	@echo "🔍 Gosec security scan..."
	@gosec -quiet ./...
	@echo "🔍 Go vet analysis..."
	@go vet ./...
	@echo "🔍 StaticCheck analysis..."
	@staticcheck ./...
	@echo "✅ Quick security scan complete"

# DAST targets (when implemented)
security-dast:
	@echo "DAST testing not yet implemented"
	@echo "Future: OAuth flow testing, API endpoint testing, etc."

# Docker Security Targets
docker-security: docker-security-scan

docker-security-scan:
	@echo "Running security scan in Docker..."
	@./scripts/docker-security.sh scan

docker-security-quick:
	@echo "Running quick security check in Docker..."
	@./scripts/docker-security.sh quick

docker-build-test:
	@echo "Testing build in Docker environment..."
	@./scripts/docker-security.sh build

docker-dev:
	@echo "Starting Docker development environment..."
	@./scripts/docker-security.sh dev

docker-stop:
	@echo "Stopping Docker services..."
	@./scripts/docker-security.sh stop

docker-cleanup:
	@echo "Cleaning up Docker resources..."
	@./scripts/docker-security.sh cleanup

# Help target
help:
	@echo "ZohoSync Makefile Commands:"
	@echo ""
	@echo "Build Targets:"
	@echo "  build              Build all applications (GUI, CLI, daemon)"
	@echo "  build-gui          Build GUI application only"
	@echo "  build-cli          Build CLI application only"
	@echo "  build-daemon       Build daemon application only"
	@echo "  clean              Clean build artifacts"
	@echo ""
	@echo "Testing Targets:"
	@echo "  test               Run Go tests"
	@echo "  lint               Run golangci-lint"
	@echo ""
	@echo "Security Targets (Local):"
	@echo "  security           Install tools + run comprehensive security scan"
	@echo "  security-install   Install security tools (gosec, staticcheck, etc.)"
	@echo "  security-scan      Run comprehensive security analysis"
	@echo "  security-quick     Run quick security checks"
	@echo ""
	@echo "Security Targets (Docker):"
	@echo "  docker-security-scan    Run comprehensive security scan in Docker"
	@echo "  docker-security-quick   Run quick security check in Docker"
	@echo "  docker-build-test       Test build in Docker environment"
	@echo "  docker-dev             Start Docker development environment"
	@echo "  docker-stop            Stop Docker services"
	@echo "  docker-cleanup         Clean up Docker resources"
	@echo ""
	@echo "Installation:"
	@echo "  install            Install binaries to /usr/local/bin"
	@echo ""
	@echo "Examples:"
	@echo "  make docker-security-scan    # Run full security analysis in Docker"
	@echo "  make docker-dev              # Start development environment"
	@echo "  make docker-build-test       # Test build without local Go"

.DEFAULT_GOAL := build
