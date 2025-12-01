# Makefile for ops-cli

# Variables
BINARY_NAME=ops-cli
MAIN_PACKAGE=main.go
BUILD_DIR=bin
PLUGINS_DIR=plugins
# Use WORKSPACE_HOME if set, otherwise use HOME
WORKSPACE_HOME ?= $(shell echo $$HOME)
PLUGIN_OUTPUT_DIR=$(WORKSPACE_HOME)/.config/ops-cli/plugins

# Version information (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
BUILD_FLAGS=-trimpath

# Go build settings
GO=go
GO_BUILD=$(GO) build $(BUILD_FLAGS) $(LDFLAGS)
GO_BUILD_PLUGIN=$(GO) build -buildmode=plugin $(BUILD_FLAGS)

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: all build clean test install plugin-install plugins help version

# Default target
all: build

# Help target
help:
	@echo "$(COLOR_BOLD)ops-cli Makefile$(COLOR_RESET)"
	@echo ""
	@echo "Available targets:"
	@echo "  $(COLOR_GREEN)build$(COLOR_RESET)          - Build the main binary"
	@echo "  $(COLOR_GREEN)plugins$(COLOR_RESET)         - Build all plugins"
	@echo "  $(COLOR_GREEN)plugin-install$(COLOR_RESET)  - Install all plugins to plugin directory"
	@echo "  $(COLOR_GREEN)test$(COLOR_RESET)            - Run tests"
	@echo "  $(COLOR_GREEN)clean$(COLOR_RESET)           - Remove build artifacts"
	@echo "  $(COLOR_GREEN)install$(COLOR_RESET)         - Install binary to system (requires sudo)"
	@echo "  $(COLOR_GREEN)version$(COLOR_RESET)         - Show version information"
	@echo ""
	@echo "Plugin targets:"
	@echo "  $(COLOR_GREEN)plugin-<name>$(COLOR_RESET) - Build specific plugin (e.g., plugin-newrelic)"
	@echo "  $(COLOR_GREEN)install-<name>$(COLOR_RESET) - Install specific plugin (e.g., install-newrelic)"
	@echo ""

# Build the main binary
build:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(COLOR_GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

# Build all plugins
plugins:
	@echo "$(COLOR_BLUE)Building all plugins...$(COLOR_RESET)"
	@for plugin in $$(find $(PLUGINS_DIR) -mindepth 1 -maxdepth 1 -type d ! -name '.*' -exec basename {} \;); do \
		echo "$(COLOR_YELLOW)Building plugin: $$plugin$(COLOR_RESET)"; \
		$(MAKE) plugin-$$plugin || exit 1; \
	done
	@echo "$(COLOR_GREEN)✓ All plugins built$(COLOR_RESET)"

# Install all plugins
plugin-install: plugins
	@echo "$(COLOR_BLUE)Installing all plugins...$(COLOR_RESET)"
	@mkdir -p $(PLUGIN_OUTPUT_DIR)
	@for plugin in $$(find $(PLUGINS_DIR) -mindepth 1 -maxdepth 1 -type d ! -name '.*' -exec basename {} \;); do \
		echo "$(COLOR_YELLOW)Installing plugin: $$plugin$(COLOR_RESET)"; \
		$(MAKE) install-$$plugin || exit 1; \
	done
	@echo "$(COLOR_GREEN)✓ All plugins installed$(COLOR_RESET)"

# Build a specific plugin
plugin-%:
	@plugin_name="$*"; \
	if [ ! -d "$(PLUGINS_DIR)/$$plugin_name" ]; then \
		echo "$(COLOR_YELLOW)Plugin $$plugin_name not found$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	echo "$(COLOR_BLUE)Building plugin: $$plugin_name$(COLOR_RESET)"; \
	plugin_files=$$(find $(PLUGINS_DIR)/$$plugin_name -name "*.go" -type f | grep -v "/commands/" | tr '\n' ' '); \
	if [ -z "$$plugin_files" ]; then \
		plugin_files=$$(find $(PLUGINS_DIR)/$$plugin_name -name "*.go" -type f | tr '\n' ' '); \
	fi; \
	if [ -z "$$plugin_files" ]; then \
		echo "$(COLOR_YELLOW)No Go files found in $(PLUGINS_DIR)/$$plugin_name$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	$(GO_BUILD_PLUGIN) -o $(BUILD_DIR)/$$plugin_name.so $$plugin_files; \
	echo "$(COLOR_GREEN)✓ Built plugin: $$plugin_name$(COLOR_RESET)"

# Install a specific plugin
install-%: plugin-%
	@plugin_name="$*"; \
	mkdir -p $(PLUGIN_OUTPUT_DIR); \
	cp $(BUILD_DIR)/$$plugin_name.so $(PLUGIN_OUTPUT_DIR)/$$plugin_name.so; \
	echo "$(COLOR_GREEN)✓ Installed plugin: $$plugin_name to $(PLUGIN_OUTPUT_DIR)$(COLOR_RESET)"

# Run tests
test:
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

# Clean build artifacts
clean:
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

# Clean plugins from plugin directory
clean-plugins:
	@echo "$(COLOR_BLUE)Cleaning installed plugins...$(COLOR_RESET)"
	rm -f $(PLUGIN_OUTPUT_DIR)/*.so
	@echo "$(COLOR_GREEN)✓ Plugins cleaned$(COLOR_RESET)"

# Install binary to system (requires sudo)
install: build
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(COLOR_RESET)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Installed $(BINARY_NAME)$(COLOR_RESET)"

# Uninstall binary from system (requires sudo)
uninstall:
	@echo "$(COLOR_BLUE)Uninstalling $(BINARY_NAME)...$(COLOR_RESET)"
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Uninstalled $(BINARY_NAME)$(COLOR_RESET)"

# Show version information
version:
	@echo "$(COLOR_BOLD)Version Information:$(COLOR_RESET)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Date:    $(DATE)"
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		echo ""; \
		echo "$(COLOR_BOLD)Binary Information:$(COLOR_RESET)"; \
		$(BUILD_DIR)/$(BINARY_NAME) --version 2>/dev/null || echo "  Binary not found or version not available"; \
	fi

# Format code
fmt:
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

# Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(COLOR_BLUE)Linting code...$(COLOR_RESET)"; \
		golangci-lint run; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)golangci-lint not found. Install with: brew install golangci-lint$(COLOR_RESET)"; \
	fi

# Run vet
vet:
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

# Build for development (faster, no optimizations)
dev: BUILD_FLAGS=-trimpath -race
dev: build

# Build for release (optimized)
release: BUILD_FLAGS=-trimpath -ldflags="-s -w"
release: build

# Rebuild everything
rebuild: clean build

# Quick development cycle: build and test
quick: fmt vet test build

