MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: apply build clean fmt help install lint quality test uninstall

BINARY_NAME := coati
BUILD_DIR   := bin
INSTALL_DIR := $(HOME)/.local/bin

help: ## Show available targets
	@echo "coati — Available targets"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the project
	@./make/build.sh

test: ## Run tests
	@./make/test.sh

apply: build ## Generate and apply hosts and SSH config to the system
	@./$(BUILD_DIR)/$(BINARY_NAME) process -f
	@./make/update.sh

clean: ## Clean build artifacts
	@./make/clean.sh

install: build ## Install the binary
	@./make/install.sh

uninstall: ## Uninstall the binary
	@./make/uninstall.sh

lint: ## Run linter
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

quality: fmt lint ## Run all quality checks

