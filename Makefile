.PHONY: run dev build test clean install-air docker-build

ifneq (,$(wildcard .env))
include .env
export
endif

# Install air for hot reload
install-air:
	@echo "Installing air..."
	@go install github.com/air-verse/air@latest
	@echo "Air installed successfully!"

# Run with hot reload (development)
dev:
	@if command -v air > /dev/null; then \
		air; \
	elif [ -f $(HOME)/go/bin/air ]; then \
		$(HOME)/go/bin/air; \
	else \
		echo "Air not found. Run 'make install-air' first."; \
	fi

# Run without hot reload
run:
	@go run ./cmd/api

# Build the application
build:
	@go build -o bin/api ./cmd/api

# Run tests
test:
	@go test ./...

# Build Docker image
docker-build:
	@docker build -t mindex-api .

# Clean build artifacts
clean:
	@rm -rf bin/ tmp/ build-errors.log
