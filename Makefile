# Crackerbox — Unified Build System
# ──────────────────────────────────

VERSION   ?= 0.1.0-dev
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS   := -s -w -X cmd.Version=$(VERSION) -X cmd.GitCommit=$(COMMIT) -X cmd.BuildDate=$(BUILD_DATE)

.PHONY: all build build-daemon build-manager build-cli clean test help

## Build all binaries
all: build

## Build all Go binaries
build: build-daemon build-manager build-cli
	@echo "✅ All binaries built successfully"

## Build crackerboxd (daemon)
build-daemon:
	@echo "🔥 Building crackerboxd..."
	cd apps/matchboxd && go build -ldflags "$(LDFLAGS)" -o ../../bin/crackerboxd .

## Build crackerbox-manager (orchestrator)
build-manager:
	@echo "🔥 Building crackerbox-manager..."
	cd apps/manager && go build -ldflags "$(LDFLAGS)" -o ../../bin/crackerbox-manager .

## Build crackerbox CLI
build-cli:
	@echo "🔥 Building crackerbox..."
	cd apps/cli && go build -ldflags "$(LDFLAGS)" -o ../../bin/crackerbox .

## Run all tests
test:
	@echo "🧪 Testing crackerboxd..."
	cd apps/matchboxd && go test -v ./...
	@echo "🧪 Testing crackerbox-manager..."
	cd apps/manager && go test -v ./...
	@echo "🧪 Testing crackerbox CLI..."
	cd apps/cli && go test -v ./...
	@echo "✅ All tests passed"

## Clean build artifacts
clean:
	rm -rf bin/
	@echo "🧹 Cleaned build artifacts"

## Show help
help:
	@echo "Crackerbox Build System"
	@echo "━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make build          Build all binaries → bin/"
	@echo "  make build-daemon   Build crackerboxd"
	@echo "  make build-manager  Build crackerbox-manager"
	@echo "  make build-cli      Build crackerbox CLI"
	@echo "  make test           Run all tests"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Outputs go to bin/ directory"
