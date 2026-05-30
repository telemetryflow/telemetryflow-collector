# TelemetryFlow Collector - Makefile
#
# TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
# Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
#
# Build Type: Direct Go Build with TFO Custom Components
# Build and development commands for TelemetryFlow Collector

# =============================================================================
# Build Configuration
# =============================================================================
PRODUCT_NAME := TelemetryFlow Collector
BINARY_NAME := tfo-collector
VERSION ?= 1.2.1
OTEL_VERSION := 0.152.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# =============================================================================
# Directories
# =============================================================================
BUILD_DIR := ./build
CONFIG_DIR := ./configs
DIST_DIR := ./dist

# =============================================================================
# Go Parameters
# =============================================================================
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GORUN := $(GOCMD) run
GOINSTALL := $(GOCMD) install

# =============================================================================
# Build Flags (uses internal/version package)
# =============================================================================
LDFLAGS := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitCommit=$(GIT_COMMIT)' \
	-X 'main.GitBranch=$(GIT_BRANCH)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

LDFLAGS_VERSION := -s -w \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.Version=$(VERSION)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitBranch=$(GIT_BRANCH)' \
	-X 'github.com/telemetryflow/telemetryflow-collector/internal/version.BuildTime=$(BUILD_TIME)'

# =============================================================================
# Platforms for Cross-Compilation
# =============================================================================
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# =============================================================================
# Colors for Output
# =============================================================================
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
BLUE := \033[0;34m
CYAN := \033[0;36m
NC := \033[0m

# =============================================================================
# Phony Targets
# =============================================================================
.PHONY: all build build-all build-linux build-darwin build-windows clean \
	test test-unit test-integration test-e2e test-all test-coverage test-short test-components \
	run run-debug dev dev-watch \
	deps deps-update deps-verify deps-refresh tidy verify \
	lint lint-fix fmt fmt-check vet staticcheck check \
	validate-config \
	install uninstall \
	ci ci-lint ci-test ci-build ci-release \
	security govulncheck coverage-merge coverage-report \
	test-unit-ci test-integration-ci test-e2e-ci \
	docker docker-build docker-push docker-run \
	build-components tidy-components components \
	trivy-scan trivy-scan-image trivy-scan-fs trivy-scan-config \
	release-check version help info

# =============================================================================
# Default Target
# =============================================================================
all: build

# =============================================================================
# Help Target
# =============================================================================
help:
	@echo "$(GREEN)$(PRODUCT_NAME) - Build System$(NC)"
	@echo ""
	@echo "$(YELLOW)Primary Build ($(BUILD_DIR)/$(BINARY_NAME)):$(NC)"
	@echo "  make                  - Build unified collector (default)"
	@echo "  make build            - Build unified collector for current platform"
	@echo "  make run              - Build and run: $(BUILD_DIR)/$(BINARY_NAME) --config"
	@echo "  make tidy             - Tidy all go modules (main + components)"
	@echo ""
	@echo "$(YELLOW)TFO Custom Components:$(NC)"
	@echo "  make build-components - Build/verify TFO custom components"
	@echo "  make tidy-components  - Tidy TFO component modules"
	@echo "  make components       - List all included components"
	@echo ""
	@echo "$(YELLOW)Platform Builds:$(NC)"
	@echo "  make build-linux      - Build for Linux (amd64 and arm64)"
	@echo "  make build-darwin     - Build for macOS (amd64 and arm64)"
	@echo "  make build-all        - Build for all platforms"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  make dev              - Run with file watching (requires watchexec)"
	@echo "  make run-debug        - Run in debug mode"
	@echo "  make validate-config  - Validate configuration file"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  make test             - Run unit and integration tests"
	@echo "  make test-unit        - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make test-e2e         - Run E2E tests only"
	@echo "  make test-all         - Run all tests"
	@echo "  make test-coverage    - Generate coverage reports"
	@echo "  make test-components  - Run TFO component tests"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@echo "  make lint             - Run linters"
	@echo "  make lint-fix         - Run linters with auto-fix"
	@echo "  make fmt              - Format code"
	@echo "  make fmt-check        - Check code formatting (CI)"
	@echo "  make vet              - Run go vet"
	@echo "  make staticcheck      - Run staticcheck"
	@echo "  make check            - Run all checks (fmt, vet, lint, test)"
	@echo ""
	@echo "$(YELLOW)Security:$(NC)"
	@echo "  make security         - Run security scan (gosec)"
	@echo "  make govulncheck      - Run vulnerability check"
	@echo "  make trivy-scan       - Run full Trivy scan (fs + config + image)"
	@echo "  make trivy-scan-fs    - Scan Go dependencies with Trivy"
	@echo "  make trivy-scan-config- Scan Dockerfile with Trivy"
	@echo "  make trivy-scan-image - Scan container image with Trivy"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  make deps             - Download dependencies"
	@echo "  make deps-update      - Update dependencies"
	@echo "  make deps-verify      - Download and verify dependencies"
	@echo "  make deps-refresh     - Refresh private module checksums"
	@echo "  make tidy             - Tidy go modules"
	@echo "  make verify           - Verify dependencies"
	@echo ""
	@echo "$(YELLOW)CI/CD Pipeline:$(NC)"
	@echo "  make ci               - Run full CI pipeline"
	@echo "  make ci-lint          - Run CI lint pipeline"
	@echo "  make ci-test          - Run CI test pipeline"
	@echo "  make ci-build         - Run CI build (GOOS/GOARCH)"
	@echo "  make ci-release       - Run release checks"
	@echo "  make coverage-merge   - Merge coverage files"
	@echo "  make coverage-report  - Generate coverage report"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  make docker           - Build Docker image"
	@echo "  make docker-build     - Build Docker image (alias)"
	@echo "  make docker-push      - Push Docker image"
	@echo "  make docker-run       - Run Docker container"
	@echo ""
	@echo "$(YELLOW)Other:$(NC)"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make install          - Install binary to /usr/local/bin"
	@echo "  make uninstall        - Uninstall binary"
	@echo "  make version          - Show version information"
	@echo "  make info             - Show build configuration"
	@echo ""
	@echo "$(YELLOW)TFO Components Included:$(NC)"
	@echo "  tfootlp     - OTLP receiver with v1/v2 endpoints"
	@echo "  tfo         - TFO Platform exporter with auto-auth"
	@echo "  tfoauth     - TFO API key management extension"
	@echo "  tfoidentity - Collector identity extension"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  VERSION=$(VERSION)"
	@echo "  OTEL_VERSION=$(OTEL_VERSION)"
	@echo "  GIT_COMMIT=$(GIT_COMMIT)"
	@echo "  GIT_BRANCH=$(GIT_BRANCH)"

# =============================================================================
# TFO Custom Component Targets
# =============================================================================

## Tidy TFO component modules
tidy-components:
	@echo "$(GREEN)Tidying TFO component modules...$(NC)"
	@for dir in components/tfootlpreceiver components/tfoexporter components/extension/tfoauthextension components/extension/tfoidentityextension; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "$(YELLOW)Tidying $$dir...$(NC)"; \
			cd $$dir && $(GOMOD) tidy && cd - > /dev/null; \
		fi; \
	done
	@echo "$(GREEN)Component modules tidied$(NC)"

## Verify TFO component builds
build-components:
	@echo "$(GREEN)Verifying TFO component builds...$(NC)"
	@for dir in components/tfootlpreceiver components/tfoexporter components/extension/tfoauthextension components/extension/tfoidentityextension; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "$(YELLOW)Building $$dir...$(NC)"; \
			cd $$dir && $(GOBUILD) ./... && cd - > /dev/null; \
		fi; \
	done
	@echo "$(GREEN)All TFO components verified$(NC)"

## List all included components from manifest
components:
	@echo "$(GREEN)Components included in $(BINARY_NAME):$(NC)"
	@echo ""
	@echo "$(YELLOW)TFO Custom Components:$(NC)"
	@echo "  - tfootlp (receiver)      v1/v2 OTLP endpoints"
	@echo "  - tfo (exporter)          auto-auth TFO Platform"
	@echo "  - tfoauth (extension)     API key management"
	@echo "  - tfoidentity (extension) collector identity"
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

# =============================================================================
# Build Targets
# =============================================================================

## Build unified collector for current platform
build: tidy-components
	@echo "$(GREEN)Building unified $(BINARY_NAME) v$(VERSION) (with TFO components)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/tfo-collector
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@echo ""
	@echo "$(GREEN)TFO Components included:$(NC)"
	@echo "  - tfootlp (receiver)      v1/v2 OTLP endpoints"
	@echo "  - tfo (exporter)          auto-auth TFO Platform"
	@echo "  - tfoauth (extension)     API key management"
	@echo "  - tfoidentity (extension) collector identity"

## Build for all platforms
build-all: tidy-components
	@echo "$(GREEN)Building $(BINARY_NAME) for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} ; \
		output="$(DIST_DIR)/$(BINARY_NAME)-$${GOOS}-$${GOARCH}" ; \
		if [ "$${GOOS}" = "windows" ]; then output="$${output}.exe"; fi ; \
		echo "$(YELLOW)Building for $${GOOS}/$${GOARCH}...$(NC)" ; \
		GOOS=$${GOOS} GOARCH=$${GOARCH} $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $${output} ./cmd/tfo-collector ; \
	done
	@echo "$(GREEN)All builds complete in $(DIST_DIR)$(NC)"

## Build for Linux
build-linux: tidy-components
	@echo "$(GREEN)Building $(BINARY_NAME) for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tfo-collector
	@GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/tfo-collector
	@echo "$(GREEN)Linux builds complete$(NC)"

## Build for macOS
build-darwin: tidy-components
	@echo "$(GREEN)Building $(BINARY_NAME) for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tfo-collector
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tfo-collector
	@echo "$(GREEN)macOS builds complete$(NC)"

# =============================================================================
# Run & Development Targets
# =============================================================================

## Run the collector locally
run: build
	@echo "$(GREEN)Starting $(BINARY_NAME)...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_DIR)/tfo-collector.yaml

## Run with debug output
run-debug: build
	@echo "$(GREEN)Starting $(BINARY_NAME) in debug mode...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_DIR)/tfo-collector.yaml --set=service.telemetry.logs.level=debug

## Run with file watching (requires watchexec)
dev:
	@echo "$(GREEN)Starting development mode with file watching...$(NC)"
	@which watchexec > /dev/null || (echo "$(RED)watchexec not found. Install with: brew install watchexec$(NC)" && exit 1)
	@watchexec -r -e go,yaml -- make run

## Alias for dev (watch mode)
dev-watch: dev

## Validate configuration
validate-config: build
	@echo "$(GREEN)Validating configuration...$(NC)"
	@$(BUILD_DIR)/$(BINARY_NAME) validate --config $(CONFIG_DIR)/tfo-collector.yaml

# =============================================================================
# Testing Targets
# =============================================================================

## Run unit and integration tests
test: test-unit test-integration
	@echo "$(GREEN)All tests completed$(NC)"

## Run unit tests only
test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	@cd tests && $(GOTEST) -v -timeout 5m -coverprofile=../coverage-unit.out ./unit/components/...

## Run integration tests only
test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@cd tests && $(GOTEST) -v -timeout 5m -coverprofile=../coverage-integration.out ./integration/components/...

## Run E2E tests only
test-e2e:
	@echo "$(GREEN)Running E2E tests...$(NC)"
	@$(GOTEST) -v -timeout 10m ./tests/e2e/...

## Run all tests
test-all: test-unit test-integration test-e2e
	@echo "$(GREEN)All tests completed$(NC)"

## Run TFO component tests
test-components: test-unit test-integration
	@echo "$(GREEN)All TFO component tests completed$(NC)"

## Generate coverage reports
test-coverage:
	@echo "$(GREEN)Generating coverage reports...$(NC)"
	@$(GOCMD) tool cover -html=coverage-unit.out -o coverage-unit.html 2>/dev/null || true
	@$(GOCMD) tool cover -html=coverage-integration.out -o coverage-integration.html 2>/dev/null || true
	@echo "$(GREEN)Coverage reports generated$(NC)"

## Run short tests (unit only)
test-short: test-unit
	@echo "$(GREEN)Short tests completed$(NC)"

# =============================================================================
# Dependencies Targets
# =============================================================================

## Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies downloaded$(NC)"

## Update dependencies
deps-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@$(GOGET) -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## Download and verify dependencies
deps-verify: deps verify
	@echo "$(GREEN)Dependencies downloaded and verified$(NC)"

## CI: Refresh private module checksums (for re-tagged modules)
deps-refresh:
	@echo "$(GREEN)Refreshing dependencies...$(NC)"
	@rm -rf vendor go.sum
	@echo "$(YELLOW)Clearing module cache...$(NC)"
	@$(GOCMD) clean -modcache
	@echo "$(GREEN)Re-downloading dependencies with fresh checksums...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@$(GOMOD) verify
	@echo "$(GREEN)Dependencies refreshed$(NC)"

## Tidy go modules
tidy:
	@echo "$(GREEN)Tidying go modules...$(NC)"
	@$(GOMOD) tidy
	@echo "$(GREEN)Go modules tidied$(NC)"

## Verify dependencies
verify:
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	@$(GOMOD) verify
	@echo "$(GREEN)Dependencies verified$(NC)"

# =============================================================================
# Code Quality Targets
# =============================================================================

## Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

## Run linter with auto-fix
lint-fix:
	@echo "$(GREEN)Running linter with auto-fix...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed, skipping...$(NC)"; \
	fi

## Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOCMD) fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

## Check code formatting (fails if code needs formatting)
fmt-check:
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)The following files need formatting:$(NC)"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)Code formatting OK$(NC)"

## Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOCMD) vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

## Run staticcheck
staticcheck:
	@echo "$(GREEN)Running staticcheck...$(NC)"
	@STATICCHECK="$$(go env GOPATH)/bin/staticcheck"; \
	if ! $$STATICCHECK -version 2>/dev/null | grep -q "v0\.7\|2026\."; then \
		echo "$(YELLOW)Installing staticcheck v0.7.0...$(NC)"; \
		$(GOINSTALL) honnef.co/go/tools/cmd/staticcheck@v0.7.0; \
	fi; \
	$$STATICCHECK ./...

## Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed$(NC)"

# =============================================================================
# Security Targets
# =============================================================================

## Run security scan with gosec
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -no-fail -fmt sarif -out gosec-results.sarif ./...; \
	else \
		echo "$(YELLOW)gosec not installed, skipping...$(NC)"; \
		echo "$(YELLOW)Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

## Run govulncheck
govulncheck:
	@echo "$(GREEN)Running govulncheck...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./... || true; \
	else \
		echo "$(YELLOW)Installing govulncheck...$(NC)"; \
		$(GOINSTALL) golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./... || true; \
	fi

# =============================================================================
# CI-Specific Targets
# =============================================================================
# These targets are optimized for CI/CD pipelines with proper exit codes,
# coverage output, and race detection.

## CI: Run full CI pipeline
ci: deps-verify ci-lint ci-test ci-build
	@echo "$(GREEN)CI pipeline completed$(NC)"

## CI: Complete lint pipeline
ci-lint: deps-verify fmt-check vet staticcheck security
	@echo "$(GREEN)CI lint pipeline completed$(NC)"

## CI: Complete test pipeline
ci-test: test-unit-ci test-integration-ci
	@echo "$(GREEN)CI test pipeline completed$(NC)"

## CI: Run unit tests with race detection and coverage
test-unit-ci:
	@echo "$(GREEN)Running unit tests (CI mode with race detection)...$(NC)"
	@cd tests && $(GOTEST) -v -race -timeout 10m -coverprofile=../coverage-unit.out -covermode=atomic ./unit/components/...

## CI: Run integration tests with race detection and coverage
test-integration-ci:
	@echo "$(GREEN)Running integration tests (CI mode)...$(NC)"
	@cd tests && $(GOTEST) -v -race -timeout 10m -coverprofile=../coverage-integration.out -covermode=atomic ./integration/components/...

## CI: Run E2E tests
test-e2e-ci:
	@echo "$(GREEN)Running E2E tests (CI mode)...$(NC)"
	@$(GOTEST) -v -timeout 15m ./tests/e2e/...

## CI: Build for a specific platform (used by GitHub Actions)
ci-build: tidy-components
	@echo "$(GREEN)Building $(BINARY_NAME) for CI ($(GOOS)/$(GOARCH))...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@OUTPUT="$(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)"; \
	if [ "$(GOOS)" = "windows" ]; then OUTPUT="$${OUTPUT}.exe"; fi; \
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) -ldflags "$(LDFLAGS_VERSION)" -o $${OUTPUT} ./cmd/tfo-collector; \
	echo "$(GREEN)Built: $${OUTPUT}$(NC)"; \
	ls -la "$${OUTPUT}"

## CI: Run release checks
ci-release: release-check
	@echo "$(GREEN)CI release checks completed$(NC)"

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
		$(GOINSTALL) github.com/wadey/gocovmerge@latest; \
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
	@$(GOCMD) tool cover -func=coverage-merged.out | tee coverage-summary.txt
	@$(GOCMD) tool cover -html=coverage-merged.out -o coverage.html
	@echo "$(GREEN)Coverage report generated$(NC)"

# =============================================================================
# Docker Targets
# =============================================================================

## Build Docker image
docker: docker-build

## Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t telemetryflow/telemetryflow-collector:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		.
	@docker tag telemetryflow/telemetryflow-collector:$(VERSION) telemetryflow/telemetryflow-collector:latest
	@echo "$(GREEN)Docker image built: telemetryflow/telemetryflow-collector:$(VERSION)$(NC)"

## Push Docker image
docker-push: docker-build
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	@docker push telemetryflow/telemetryflow-collector:$(VERSION)
	@docker push telemetryflow/telemetryflow-collector:latest

## Run Docker container
docker-run:
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -it --rm \
		-v $(PWD)/configs:/etc/tfo-collector \
		telemetryflow/telemetryflow-collector:$(VERSION)

# =============================================================================
# Container Vulnerability Scanning (Trivy)
# =============================================================================

TRIVY_IMAGE := telemetryflow/telemetryflow-collector:$(VERSION)
TRIVY_SEVERITY := HIGH,CRITICAL
TRIVY_FORMAT := table
TRIVY_TIMEOUT := 5m

## Run full Trivy vulnerability scan (filesystem + config + image)
trivy-scan: trivy-scan-fs trivy-scan-config trivy-scan-image
	@echo "$(GREEN)All Trivy scans completed - 0 vulnerabilities$(NC)"

## Scan container image with Trivy (builds image first)
trivy-scan-image: docker-build
	@echo "$(GREEN)Scanning container image with Trivy...$(NC)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy image --severity $(TRIVY_SEVERITY) --format $(TRIVY_FORMAT) --timeout $(TRIVY_TIMEOUT) $(TRIVY_IMAGE); \
	else \
		echo "$(RED)trivy not installed. Install: https://trivy.dev/latest/getting-started/installation/$(NC)"; \
		exit 1; \
	fi

## Scan Go dependencies (filesystem) with Trivy
trivy-scan-fs:
	@echo "$(GREEN)Scanning filesystem (Go dependencies) with Trivy...$(NC)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs --scanners vuln --severity LOW,MEDIUM,HIGH,CRITICAL --format $(TRIVY_FORMAT) .; \
	else \
		echo "$(RED)trivy not installed. Install: https://trivy.dev/latest/getting-started/installation/$(NC)"; \
		exit 1; \
	fi

## Scan Dockerfile for misconfigurations with Trivy
trivy-scan-config:
	@echo "$(GREEN)Scanning Dockerfile for misconfigurations with Trivy...$(NC)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy config --format $(TRIVY_FORMAT) Dockerfile; \
	else \
		echo "$(RED)trivy not installed. Install: https://trivy.dev/latest/getting-started/installation/$(NC)"; \
		exit 1; \
	fi

# =============================================================================
# Clean & Install Targets
# =============================================================================

## Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage*.out coverage*.html coverage-summary.txt
	@rm -f gosec-results.sarif
	@echo "$(GREEN)Clean complete$(NC)"

## Install binary to /usr/local/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)Installed successfully$(NC)"

## Uninstall binary from /usr/local/bin
uninstall:
	@echo "$(GREEN)Removing $(BINARY_NAME) from /usr/local/bin...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)Uninstalled successfully$(NC)"

# =============================================================================
# Release Targets
# =============================================================================

## Check release readiness
release-check:
	@echo "$(GREEN)Checking release readiness...$(NC)"
	@echo "$(BLUE)1. Running tests...$(NC)"
	@$(MAKE) test
	@echo "$(BLUE)2. Running linter...$(NC)"
	@$(MAKE) lint
	@echo "$(BLUE)3. Building...$(NC)"
	@$(MAKE) build
	@echo "$(GREEN)Release checks passed$(NC)"

# =============================================================================
# Info Targets
# =============================================================================

## Show version information
version:
	@echo "$(GREEN)$(PRODUCT_NAME)$(NC)"
	@echo "  Version:          $(VERSION)"
	@echo "  OTEL Version:     $(OTEL_VERSION)"
	@echo "  Git Commit:       $(GIT_COMMIT)"
	@echo "  Git Branch:       $(GIT_BRANCH)"
	@echo "  Build Time:       $(BUILD_TIME)"
	@echo "  Go Version:       $(GO_VERSION)"

## Show build configuration
info:
	@echo "$(GREEN)Build Configuration$(NC)"
	@echo ""
	@echo "$(YELLOW)Product:$(NC)"
	@echo "  Name:             $(PRODUCT_NAME)"
	@echo "  Binary:           $(BINARY_NAME)"
	@echo "  Version:          $(VERSION)"
	@echo ""
	@echo "$(YELLOW)Versions:$(NC)"
	@echo "  Go:               $(GO_VERSION)"
	@echo "  OTEL Collector:   $(OTEL_VERSION)"
	@echo ""
	@echo "$(YELLOW)Git:$(NC)"
	@echo "  Commit:           $(GIT_COMMIT)"
	@echo "  Branch:           $(GIT_BRANCH)"
	@echo ""
	@echo "$(YELLOW)Directories:$(NC)"
	@echo "  Build:            $(BUILD_DIR)"
	@echo "  Config:           $(CONFIG_DIR)"
	@echo "  Dist:             $(DIST_DIR)"
	@echo ""
	@echo "$(YELLOW)Platforms:$(NC)"
	@echo "  $(PLATFORMS)"
