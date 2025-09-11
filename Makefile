MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean deb fmt help install lint package quality test uninstall

BINARY_NAME := coati

export BINARY_NAME

help: ## Show available targets
	@echo "coati - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@./.make/build.sh

test: ## Run tests
	@./.make/test.sh

lint: ## Run linter
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

quality: fmt lint ## Run all quality checks

clean: ## Remove build artifacts
	@go clean
	@rm -rf bin/

install: ## Install to ~/.local/bin (requires prior make build)
	@./.make/install.sh

uninstall: ## Remove from ~/.local/bin (sudo only if needed)
	@./.make/uninstall.sh

package: ## Build a .deb into dist/ (requires prior make build)
	@./.make/package.sh

deb: package ## Alias for package
