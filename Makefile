# Makefile for Husky Project

.PHONY: all build clean run migrate test docker-up docker-down help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=husky-api
MIGRATE_NAME=husky-migrate

# Build directories
BUILD_DIR=bin

all: build

## web-build: Build Vue frontend (embed mode) and copy to admin/static
web-build:
	@echo "Building Vue admin frontend..."
	cd web && npm ci && npx vite build --mode embed
	@echo "Copying to internal/admin/static..."
	rm -rf internal/admin/static
	cp -r web/dist internal/admin/static
	@echo "Renaming _-prefixed files (Go embed excludes them)..."
	cd internal/admin/static && find . -name '_*' -type f | while read f; do \
		dir=$$(dirname "$$f"); \
		base=$$(basename "$$f"); \
		newbase="x$${base}"; \
		mv "$$dir/$$base" "$$dir/$$newbase"; \
	done
	@echo "Updating references to renamed files..."
	cd internal/admin/static && \
		find assets -name '*.js' -exec sed -i 's|_plugin-vue_export-helper|x_plugin-vue_export-helper|g' {} + && \
		find assets -name '*.css' -exec sed -i 's|_plugin-vue_export-helper|x_plugin-vue_export-helper|g' {} + && \
		find . -name '*.html' -exec sed -i 's|_plugin-vue_export-helper|x_plugin-vue_export-helper|g' {} +
	@echo "Web build completed!"

## build: Build both API and migrate binaries (runs web-build first)
build: web-build
	@echo "Building Husky API..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "Building Husky Migrate..."
	$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_NAME) ./cmd/migrate
	@echo "Build completed!"

## build-go: Build Go binaries only (skip frontend build)
build-go:
	@echo "Building Husky API..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api
	@echo "Building Husky Migrate..."
	$(GOBUILD) -o $(BUILD_DIR)/$(MIGRATE_NAME) ./cmd/migrate
	@echo "Build completed!"

## run: Run the API server
run:
	@echo "Starting Husky API server..."
	$(GORUN) ./cmd/api

## migrate: Run database migrations
migrate:
	@echo "Running database migrations..."
	$(GORUN) ./cmd/migrate

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/*
	rm -f coverage.out coverage.html
	@echo "Clean completed!"

## docker-up: Start Docker containers
docker-up:
	@echo "Starting Docker containers..."
	docker-compose up -d
	@echo "Containers started!"

## docker-down: Stop Docker containers
docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down
	@echo "Containers stopped!"

## docker-logs: View Docker container logs
docker-logs:
	docker-compose logs -f

## docker-restart: Restart Docker containers
docker-restart: docker-down docker-up

## dev: Run in development mode with hot reload (requires air)
dev:
	@echo "Starting development mode..."
	air -c .air.toml

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

## fmt: Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .

## help: Show this help message
help:
	@echo "Husky Project - Enterprise Intelligent Ticket System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
