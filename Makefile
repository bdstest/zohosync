# ZohoSync Makefile

APP_NAME := zohosync
CLI_NAME := $(APP_NAME)-cli
DAEMON_NAME := $(APP_NAME)-daemon

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.commit=$(COMMIT)

.PHONY: all build clean test lint install

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
	go test -v ./...

lint:
	@echo "Running linter..."
	golangci-lint run

install: build
	@echo "Installing..."
	sudo cp $(APP_NAME) /usr/local/bin/
	sudo cp $(CLI_NAME) /usr/local/bin/
	sudo cp $(DAEMON_NAME) /usr/local/bin/

.DEFAULT_GOAL := build
