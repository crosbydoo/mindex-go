.PHONY: run run-plain dev build test clean install-air docker-build

ifneq (,$(wildcard .env))
include .env
export
endif

AIR := $(shell command -v air 2>/dev/null)
ifeq ($(AIR),)
  AIR := $(shell go env GOPATH)/bin/air
endif

# Install air for hot reload (one time)
install-air:
	@echo "Installing air..."
	@go install github.com/air-verse/air@latest
	@echo "Air installed at: $$(go env GOPATH)/bin/air"

# Run with hot reload (default)
run: ensure-air
	@$(AIR)

# Alias for hot reload
dev: run

# Run without hot reload
run-plain:
	@go run ./cmd/api

ensure-air:
	@if [ ! -x "$(AIR)" ]; then \
		echo "Air not found. Installing..."; \
		go install github.com/air-verse/air@latest; \
	fi

# Build the application
build:
	@go build -o bin/api ./cmd/api

# Run tests
test:
	@go test ./...

# Build Docker image
docker-build:
	@docker build -t ristudev/mindex-go-server .

# Clean build artifacts
clean:
	@rm -rf bin/ tmp/ build-errors.log
