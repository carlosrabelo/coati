MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help lint quality test

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
	@./.make/clean.sh
