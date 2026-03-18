# Crackerbox — Unified Build System
# ──────────────────────────────────

VERSION   ?= 0.1.0-dev
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

DAEMON_LDFLAGS  := -s -w -X crackerboxd/cmd.Version=$(VERSION) -X crackerboxd/cmd.GitCommit=$(COMMIT) -X crackerboxd/cmd.BuildDate=$(BUILD_DATE)
MANAGER_LDFLAGS := -s -w -X github.com/devsprithvi/crakerbox/apps/manager/cmd.Version=$(VERSION) -X github.com/devsprithvi/crakerbox/apps/manager/cmd.GitCommit=$(COMMIT) -X github.com/devsprithvi/crakerbox/apps/manager/cmd.BuildDate=$(BUILD_DATE)

.PHONY: all build build-daemon build-manager clean test help

## Build all binaries
all: build

## Build all Go binaries
build: build-daemon build-manager
	@echo "✅ All binaries built successfully"

## Build crackerboxd (daemon + node CLI)
build-daemon:
	@echo "🔥 Building crackerboxd..."
	cd apps/matchboxd && go build -ldflags "$(DAEMON_LDFLAGS)" -o ../../bin/crackerboxd .

## Build crackerbox-manager (orchestrator + cluster CLI)
build-manager:
	@echo "🔥 Building crackerbox-manager..."
	cd apps/manager && go build -ldflags "$(MANAGER_LDFLAGS)" -o ../../bin/crackerbox-manager .

## Run all tests
test:
	@echo "🧪 Testing crackerboxd..."
	cd apps/matchboxd && go test -v ./...
	@echo "🧪 Testing crackerbox-manager..."
	cd apps/manager && go test -v ./...
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
	@echo "  make build-daemon   Build crackerboxd (daemon + node CLI)"
	@echo "  make build-manager  Build crackerbox-manager (orchestrator + cluster CLI)"
	@echo "  make test           Run all tests"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Outputs go to bin/ directory"
