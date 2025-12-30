# TelemetryFlow Collector - Makefile
#
# TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
# Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
#
# Build and development commands for TelemetryFlow Collector

# Build configuration
PRODUCT_NAME := TelemetryFlow Collector
BINARY_NAME := tfo-collector
BINARY_NAME_OCB := tfo-collector-ocb
VERSION ?= 1.1.1
OTEL_VERSION := 0.142.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# Directories
BUILD_DIR := ./build
BUILD_DIR_OCB := ./build/ocb
CONFIG_DIR := ./configs
DIST_DIR := ./dist

# OCB (OpenTelemetry Collector Builder) - binary name is "builder"
# Try to find builder in PATH first
OCB := $(shell which builder 2>/dev/null || echo "")
ifeq ($(OCB),)
  # Try GVM package sets
  OCB := $(shell ls $(HOME)/.gvm/pkgsets/*/global/bin/builder 2>/dev/null | head -1)
endif
ifeq ($(OCB),)
  # Try GOPATH/bin (from go env)
  OCB := $(shell go env GOPATH 2>/dev/null)/bin/builder
endif
ifeq ($(OCB),)
  # Fallback to HOME/go/bin
  OCB := $(HOME)/go/bin/builder
endif
OCB_VERSION := 0.142.0

# Go build flags (for main package variables)
LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.GitBranch=$(GIT_BRANCH)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

.PHONY: all build build-all build-linux build-darwin clean test test-unit test-integration test-e2e test-all test-coverage test-script test-short lint lint-fix fmt vet help install-ocb generate run docker build-standalone run-standalone test-standalone tidy deps deps-update install uninstall validate-config \
	fmt-check staticcheck verify deps-verify test-unit-ci test-integration-ci test-e2e-ci security govulncheck coverage-merge coverage-report ci-lint ci-test ci-build-standalone ci-build-ocb

# Default target: build standalone (uses internal packages directly)
all: build-standalone

# Help target
help:
	@echo "$(GREEN)$(PRODUCT_NAME) - Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Binaries (all in $(BUILD_DIR)/):$(NC)"
	@echo "  $(BINARY_NAME)        - Standalone CLI (Cobra commands)"
	@echo "  $(BINARY_NAME_OCB)    - OCB build (standard OTel Collector)"
	@echo ""
	@echo "$(YELLOW)Standalone Build ($(BUILD_DIR)/$(BINARY_NAME)):$(NC)"
	@echo "  make                  - Build standalone collector (default)"
	@echo "  make build-standalone - Build standalone collector"
	@echo "  make run-standalone   - Run: $(BUILD_DIR)/$(BINARY_NAME) start --config configs/tfo-collector.yaml"
	@echo "  make test-standalone  - Run standalone tests"
	@echo "  make tidy             - Tidy go modules"
	@echo ""
	@echo "$(YELLOW)OCB Build ($(BUILD_DIR)/$(BINARY_NAME_OCB)):$(NC)"
	@echo "  make build-ocb        - Build OCB collector"
	@echo "  make build-all        - Build OCB for all platforms"
	@echo "  make install-ocb      - Install OpenTelemetry Collector Builder"
	@echo "  make generate         - Generate collector code using OCB"
	@echo "  make run              - Run: $(BUILD_DIR)/$(BINARY_NAME_OCB) --config configs/otel-collector.yaml"
	@echo ""
	@echo "$(YELLOW)Platform Builds:$(NC)"
	@echo "  make build-linux      - Build for Linux (amd64 and arm64)"
	@echo "  make build-darwin     - Build for macOS (amd64 and arm64)"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  make test             - Run unit and integration tests"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-e2e         - Run E2E tests only"
	@echo "  make test-all         - Run all tests"
	@echo "  make test-coverage    - Generate coverage reports"
	@echo "  make test-script      - Run test script"
	@echo "  make test-short       - Run short tests"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@echo "  make lint             - Run linters"
	@echo "  make lint-fix         - Run linters with auto-fix"
	@echo "  make fmt              - Format code"
	@echo "  make vet              - Run go vet"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  make deps             - Download dependencies"
	@echo "  make deps-update      - Update dependencies"
	@echo "  make tidy             - Tidy go modules"
	@echo ""
	@echo "$(YELLOW)Other:$(NC)"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make install          - Install binary to /usr/local/bin"
	@echo "  make uninstall        - Uninstall binary"
	@echo "  make validate-config  - Validate configuration file"
	@echo "  make docker           - Build Docker image"
	@echo "  make version          - Show version information"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  VERSION=$(VERSION)"
	@echo "  OTEL_VERSION=$(OTEL_VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
	@echo "  GIT_BRANCH=$(GIT_BRANCH)"

# Install OCB (OpenTelemetry Collector Builder)
install-ocb:
	@echo "$(GREEN)Installing OpenTelemetry Collector Builder v$(OCB_VERSION)...$(NC)"
	@go install go.opentelemetry.io/collector/cmd/builder@v$(OCB_VERSION)
	@GOBIN=$$(go env GOPATH)/bin; \
	if [ -f "$$GOBIN/builder" ]; then \
		echo "$(GREEN)Builder installed successfully at $$GOBIN/builder$(NC)"; \
	else \
		echo "$(RED)Builder installation failed - binary not found at $$GOBIN/builder$(NC)"; \
		exit 1; \
	fi

# Helper function to find builder binary at runtime
# This is needed because the Makefile variables are evaluated at parse time
define FIND_BUILDER
$(shell which builder 2>/dev/null || \
	([ -f "$$(go env GOPATH)/bin/builder" ] && echo "$$(go env GOPATH)/bin/builder") || \
	([ -f "$(HOME)/go/bin/builder" ] && echo "$(HOME)/go/bin/builder") || \
	echo "")
endef

# Check if OCB is installed
check-ocb:
	@export PATH="$$(go env GOPATH)/bin:$(HOME)/go/bin:$$PATH"; \
	BUILDER=$$(which builder 2>/dev/null); \
	if [ -z "$$BUILDER" ] || [ ! -f "$$BUILDER" ]; then \
		BUILDER="$$(go env GOPATH)/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		BUILDER="$(HOME)/go/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		echo "$(YELLOW)Builder not found. Installing...$(NC)"; \
		$(MAKE) install-ocb; \
	else \
		echo "$(GREEN)Builder found at: $$BUILDER$(NC)"; \
	fi

# Generate collector code using OCB
generate: check-ocb
	@echo "$(GREEN)Generating collector code...$(NC)"
	@mkdir -p $(BUILD_DIR_OCB)
	@export PATH="$$(go env GOPATH)/bin:$(HOME)/go/bin:$$PATH"; \
	BUILDER=$$(which builder 2>/dev/null); \
	if [ -z "$$BUILDER" ] || [ ! -f "$$BUILDER" ]; then \
		BUILDER="$$(go env GOPATH)/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		BUILDER="$(HOME)/go/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		echo "$(RED)Builder not found. Please run 'make install-ocb' first.$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Using builder at: $$BUILDER$(NC)"; \
	$$BUILDER --config manifest.yaml
	@echo "$(GREEN)Collector code generated in $(BUILD_DIR_OCB)$(NC)"

# Build the collector using OCB (uses OCB-generated main.go)
build-ocb: generate
	@echo "$(GREEN)Building $(BINARY_NAME_OCB) v$(VERSION) with OCB...$(NC)"
	@cd $(BUILD_DIR_OCB) && go build -ldflags "$(LDFLAGS)" -o ../$(BINARY_NAME_OCB) .
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME_OCB)$(NC)"

# Build for all platforms using OCB
build-all: generate
	@echo "$(GREEN)Building $(BINARY_NAME_OCB) for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} ; \
		output="$(DIST_DIR)/$(BINARY_NAME_OCB)-$${GOOS}-$${GOARCH}" ; \
		if [ "$${GOOS}" = "windows" ]; then output="$${output}.exe"; fi ; \
		echo "$(YELLOW)Building for $${GOOS}/$${GOARCH}...$(NC)" ; \
		cd $(BUILD_DIR_OCB) && GOOS=$${GOOS} GOARCH=$${GOARCH} go build -ldflags "$(LDFLAGS)" -o ../../$${output} . ; \
	done
	@echo "$(GREEN)All builds complete in $(DIST_DIR)$(NC)"

# Build standalone for Linux
build-linux:
	@echo "$(GREEN)Building $(BINARY_NAME) for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_STANDALONE)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tfo-collector
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS_STANDALONE)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/tfo-collector
	@echo "$(GREEN)Linux builds complete$(NC)"

# Build standalone for macOS
build-darwin:
	@echo "$(GREEN)Building $(BINARY_NAME) for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS_STANDALONE)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tfo-collector
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS_STANDALONE)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tfo-collector
	@echo "$(GREEN)macOS builds complete$(NC)"

# Run the OCB collector locally
run: build
	@echo "$(GREEN)Starting $(BINARY_NAME_OCB)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME_OCB) --config $(CONFIG_DIR)/otel-collector.yaml

# Run OCB with debug output
run-debug: build
	@echo "$(GREEN)Starting $(BINARY_NAME_OCB) in debug mode...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME_OCB) --config $(CONFIG_DIR)/otel-collector.yaml --set=service.telemetry.logs.level=debug

# Validate OCB configuration
validate-config-ocb: build
	@echo "$(GREEN)Validating OCB configuration...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME_OCB) validate --config $(CONFIG_DIR)/otel-collector.yaml

# Test targets
test: test-unit test-integration
	@echo "$(GREEN)All tests completed$(NC)"

test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v -timeout 5m -coverprofile=coverage-unit.out ./tests/unit/...

test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@go test -v -timeout 5m -coverprofile=coverage-integration.out ./tests/integration/...

test-e2e: build-standalone
	@echo "$(GREEN)Running E2E tests...$(NC)"
	@TFO_COLLECTOR_BINARY=$(BUILD_DIR)/$(BINARY_NAME) go test -v -timeout 10m ./tests/e2e/...

test-all: test-unit test-integration test-e2e
	@echo "$(GREEN)All tests completed$(NC)"

test-coverage:
	@echo "$(GREEN)Generating coverage reports...$(NC)"
	@go tool cover -html=coverage-unit.out -o coverage-unit.html
	@go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "$(GREEN)Coverage reports generated$(NC)"

test-script:
	@echo "$(GREEN)Running test script...$(NC)"
	@./scripts/test.sh

test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	@./scripts/test.sh short

# Code quality
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

lint-fix:
	@echo "$(GREEN)Running linters with auto-fix...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...

vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

# Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)/*
	@rm -rf $(DIST_DIR)
	@rm -f coverage-*.out coverage-*.html
	@echo "$(GREEN)Clean complete$(NC)"

# Show version information
version:
	@echo "$(GREEN)$(PRODUCT_NAME)$(NC)"
	@echo "  Version:      $(VERSION)"
	@echo "  OTEL Version: $(OTEL_VERSION)"
	@echo "  Git Commit:   $(GIT_COMMIT)"
	@echo "  Git Branch:   $(GIT_BRANCH)"
	@echo "  Build Time:   $(BUILD_TIME)"
	@echo "  Go Version:   $(GO_VERSION)"

# Build Docker image
docker:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t telemetryflow/telemetryflow-collector:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		.
	@docker tag telemetryflow/telemetryflow-collector:$(VERSION) telemetryflow/telemetryflow-collector:latest
	@echo "$(GREEN)Docker image built: telemetryflow/telemetryflow-collector:$(VERSION)$(NC)"

# Push Docker image
docker-push: docker
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	@docker push telemetryflow/telemetryflow-collector:$(VERSION)
	@docker push telemetryflow/telemetryflow-collector:latest

# Development: watch and rebuild
dev:
	@echo "$(GREEN)Starting development mode...$(NC)"
	@which watchexec > /dev/null || (echo "$(RED)watchexec not found. Install with: brew install watchexec$(NC)" && exit 1)
	@watchexec -r -e go,yaml -- make run

# Print component list from manifest
components:
	@echo "$(GREEN)Components included in $(BINARY_NAME):$(NC)"
	@echo ""
	@echo "$(YELLOW)Extensions:$(NC)"
	@grep -A 100 "^extensions:" manifest.yaml | grep "gomod:" | sed 's/.*gomod: /  - /' | head -20
	@echo ""
	@echo "$(YELLOW)Receivers:$(NC)"
	@grep -A 100 "^receivers:" manifest.yaml | grep "gomod:" | sed 's/.*gomod: /  - /' | head -20
	@echo ""
	@echo "$(YELLOW)Processors:$(NC)"
	@grep -A 100 "^processors:" manifest.yaml | grep "gomod:" | sed 's/.*gomod: /  - /' | head -20
	@echo ""
	@echo "$(YELLOW)Exporters:$(NC)"
	@grep -A 100 "^exporters:" manifest.yaml | grep "gomod:" | sed 's/.*gomod: /  - /' | head -20

# LDFLAGS for standalone build (uses internal/version package)
LDFLAGS_STANDALONE := -s -w \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.Version=$(VERSION)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitBranch=$(GIT_BRANCH)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.BuildTime=$(BUILD_TIME)'

# Build standalone version (without OCB)
build-standalone:
	@echo "$(GREEN)Building standalone $(BINARY_NAME) v$(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "$(LDFLAGS_STANDALONE)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tfo-collector
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Run standalone version
run-standalone: build-standalone
	@echo "$(GREEN)Starting standalone $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) start --config $(CONFIG_DIR)/tfo-collector.yaml

# Run standalone tests
test-standalone:
	@echo "$(GREEN)Running standalone tests...$(NC)"
	@go test -v ./...

# Dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies downloaded$(NC)"

deps-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

tidy:
	@echo "$(GREEN)Tidying go modules...$(NC)"
	@go mod tidy
	@echo "$(GREEN)Go modules tidied$(NC)"

# Installation
install: build-standalone
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Installed successfully$(NC)"

uninstall:
	@echo "$(GREEN)Removing $(BINARY_NAME) from /usr/local/bin...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)Uninstalled successfully$(NC)"

# Configuration
validate-config: build-standalone
	@echo "$(GREEN)Validating configuration...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) validate --config $(CONFIG_DIR)/tfo-collector.yaml

# =============================================================================
# CI-Specific Targets
# =============================================================================
# These targets are optimized for CI/CD pipelines with proper exit codes,
# coverage output, and race detection.

## CI: Check formatting (fails if code needs formatting)
fmt-check:
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)The following files need formatting:$(NC)"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)Code formatting OK$(NC)"

## CI: Run staticcheck
staticcheck:
	@echo "$(GREEN)Running staticcheck...$(NC)"
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "$(YELLOW)Installing staticcheck...$(NC)"; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
		staticcheck ./...; \
	fi

## CI: Verify dependencies
verify:
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	@go mod verify
	@echo "$(GREEN)Dependencies verified$(NC)"

## CI: Download and verify dependencies
deps-verify: deps verify
	@echo "$(GREEN)Dependencies downloaded and verified$(NC)"

## CI: Run unit tests with race detection and coverage
test-unit-ci:
	@echo "$(GREEN)Running unit tests (CI mode with race detection)...$(NC)"
	@go test -v -race -timeout 10m -coverprofile=coverage-unit.out -covermode=atomic ./tests/unit/...

## CI: Run integration tests with race detection and coverage
test-integration-ci:
	@echo "$(GREEN)Running integration tests (CI mode)...$(NC)"
	@go test -v -race -timeout 10m -coverprofile=coverage-integration.out -covermode=atomic ./tests/integration/...

## CI: Run E2E tests
test-e2e-ci: build-standalone
	@echo "$(GREEN)Running E2E tests (CI mode)...$(NC)"
	@TFO_COLLECTOR_BINARY=$(BUILD_DIR)/$(BINARY_NAME) go test -v -timeout 15m ./tests/e2e/...

## CI: Run security scan with gosec
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -no-fail -fmt sarif -out gosec-results.sarif ./...; \
	else \
		echo "$(YELLOW)gosec not installed, skipping...$(NC)"; \
	fi

## CI: Run govulncheck
govulncheck:
	@echo "$(GREEN)Running govulncheck...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./... || true; \
	else \
		echo "$(YELLOW)Installing govulncheck...$(NC)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./... || true; \
	fi

## CI: Merge coverage files
coverage-merge:
	@echo "$(GREEN)Merging coverage files...$(NC)"
	@if command -v gocovmerge >/dev/null 2>&1; then \
		if [ -f coverage-integration.out ]; then \
			gocovmerge coverage-unit.out coverage-integration.out > coverage-merged.out; \
		else \
			cp coverage-unit.out coverage-merged.out; \
		fi; \
	else \
		echo "$(YELLOW)Installing gocovmerge...$(NC)"; \
		go install github.com/wadey/gocovmerge@latest; \
		if [ -f coverage-integration.out ]; then \
			gocovmerge coverage-unit.out coverage-integration.out > coverage-merged.out; \
		else \
			cp coverage-unit.out coverage-merged.out; \
		fi; \
	fi
	@echo "$(GREEN)Coverage merged to coverage-merged.out$(NC)"

## CI: Generate coverage report
coverage-report: coverage-merge
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@go tool cover -func=coverage-merged.out | tee coverage-summary.txt
	@go tool cover -html=coverage-merged.out -o coverage.html
	@echo "$(GREEN)Coverage report generated$(NC)"

## CI: Complete lint pipeline
ci-lint: deps-verify fmt-check vet staticcheck lint
	@echo "$(GREEN)CI lint pipeline completed$(NC)"

## CI: Complete test pipeline
ci-test: test-unit-ci test-integration-ci
	@echo "$(GREEN)CI test pipeline completed$(NC)"

## CI: Build standalone for a specific platform
ci-build-standalone:
	@echo "$(GREEN)Building standalone for CI ($(GOOS)/$(GOARCH))...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@OUTPUT="$(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)"; \
	if [ "$(GOOS)" = "windows" ]; then OUTPUT="$${OUTPUT}.exe"; fi; \
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS_STANDALONE)" -o $${OUTPUT} ./cmd/tfo-collector; \
	echo "$(GREEN)Built: $${OUTPUT}$(NC)"

## CI: Build OCB for a specific platform (requires OCB to be installed and code generated)
ci-build-ocb: check-ocb
	@echo "$(GREEN)Building OCB for CI ($(GOOS)/$(GOARCH))...$(NC)"
	@mkdir -p $(BUILD_DIR_OCB)
	@export PATH="$$(go env GOPATH)/bin:$(HOME)/go/bin:$$PATH"; \
	BUILDER=$$(which builder 2>/dev/null); \
	if [ -z "$$BUILDER" ] || [ ! -f "$$BUILDER" ]; then \
		BUILDER="$$(go env GOPATH)/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		BUILDER="$(HOME)/go/bin/builder"; \
	fi; \
	if [ ! -f "$$BUILDER" ]; then \
		echo "$(RED)Builder not found. Please run 'make install-ocb' first.$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)Using builder at: $$BUILDER$(NC)"; \
	$$BUILDER --config manifest.yaml
	@OUTPUT="$(BUILD_DIR)/$(BINARY_NAME_OCB)-$(GOOS)-$(GOARCH)"; \
	if [ "$(GOOS)" = "windows" ]; then OUTPUT="$${OUTPUT}.exe"; fi; \
	cd $(BUILD_DIR_OCB) && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o ../$$(basename $${OUTPUT}) .; \
	echo "$(GREEN)Built: $${OUTPUT}$(NC)"
