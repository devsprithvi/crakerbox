# Matchbox — Unified Build System
# ─────────────────────────────────

VERSION    ?= 0.1.0-dev
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")

# ── Component directories ────────────────────────────────────────────
AGENT_DIR      := apps/matchbox-agent
MANAGER_DIR    := apps/matchbox-manager
CONTROLLER_DIR := apps/matchbox-controller
APISERVER_DIR  := apps/matchbox-apiserver
DATABASE_DIR   := apps/matchbox-database

# ── Linker flags ─────────────────────────────────────────────────────
LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.GitCommit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)

.PHONY: all build clean test help \
	build-agent build-manager build-controller build-apiserver build-database

## Build all binaries
all: build

## Build all Go binaries
build: build-agent build-manager build-controller build-apiserver build-database
	@echo "✅ All binaries built successfully"

## Build matchbox-agent (worker node)
build-agent:
	@echo "🔧 Building matchbox-agent..."
	cd $(AGENT_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../bin/matchbox-agent .

## Build matchbox-manager (cluster lifecycle)
build-manager:
	@echo "🔧 Building matchbox-manager..."
	cd $(MANAGER_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../bin/matchbox-manager .

## Build matchbox-controller (cluster access & control)
build-controller:
	@echo "🔧 Building matchbox-controller..."
	cd $(CONTROLLER_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../bin/matchbox-controller .

## Build matchbox-apiserver (cluster-internal API server)
build-apiserver:
	@echo "🔧 Building matchbox-apiserver..."
	cd $(APISERVER_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../bin/matchbox-apiserver .

## Build matchbox-database (cluster-internal database)
build-database:
	@echo "🔧 Building matchbox-database..."
	cd $(DATABASE_DIR) && go build -ldflags "$(LDFLAGS)" -o ../../bin/matchbox-database .

## Run all tests
test:
	@echo "🧪 Testing matchbox-agent..."
	cd $(AGENT_DIR) && go test -v ./...
	@echo "🧪 Testing matchbox-manager..."
	cd $(MANAGER_DIR) && go test -v ./...
	@echo "🧪 Testing matchbox-controller..."
	cd $(CONTROLLER_DIR) && go test -v ./...
	@echo "🧪 Testing matchbox-apiserver..."
	cd $(APISERVER_DIR) && go test -v ./...
	@echo "🧪 Testing matchbox-database..."
	cd $(DATABASE_DIR) && go test -v ./...
	@echo "✅ All tests passed"

## Clean build artifacts
clean:
	rm -rf bin/
	@echo "🧹 Cleaned build artifacts"

## Show help
help:
	@echo "Matchbox Build System"
	@echo "━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "  make build              Build all binaries → bin/"
	@echo ""
	@echo "  make build-agent        Build matchbox-agent       (worker node)"
	@echo "  make build-manager      Build matchbox-manager     (cluster lifecycle)"
	@echo "  make build-controller   Build matchbox-controller  (cluster access)"
	@echo "  make build-apiserver    Build matchbox-apiserver   (cluster-internal)"
	@echo "  make build-database     Build matchbox-database    (cluster-internal)"
	@echo ""
	@echo "  make test               Run all tests"
	@echo "  make clean              Remove build artifacts"
	@echo ""
	@echo "Outputs go to bin/ directory"
