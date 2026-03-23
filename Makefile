MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help install lint quality run test uninstall update

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

run: build ## Run the project
	@./$(BUILD_DIR)/$(BINARY_NAME)

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

update: run ## Update hosts and SSH config
	@./make/update.sh
