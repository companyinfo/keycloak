.PHONY: help test test-unit test-integration test-coverage test-all fmt vet lint clean setup-keycloak start-keycloak stop-keycloak

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Testing targets
test: test-unit ## Run unit tests (default)

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short -race ./...

test-coverage: ## Run unit tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -short -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "Generate HTML report with: make coverage-html"

coverage-html: coverage.out ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests (requires Keycloak)
	@echo "Running integration tests..."
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'make setup-keycloak' first."; \
		exit 1; \
	fi
	@set -a && . ./.env && set +a && go test -v -race -tags=integration -coverprofile=coverage-integration.out ./...

test-all: test-unit test-integration ## Run all tests (unit + integration)

# Code quality targets
fmt: ## Format code with gofmt
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

# Keycloak setup targets
start-keycloak: ## Start Keycloak with Docker Compose
	@echo "Starting Keycloak..."
	docker compose up -d
	@echo "Waiting for Keycloak to be ready..."
	@sleep 5
	@echo "Keycloak is starting. Check status with: docker compose logs -f keycloak"

stop-keycloak: ## Stop Keycloak
	@echo "Stopping Keycloak..."
	docker compose down

setup-keycloak: start-keycloak ## Setup Keycloak for integration testing
	@echo "Waiting for Keycloak to be fully ready (this may take 60-90 seconds)..."
	@echo "Checking Keycloak status..."
	@for i in $$(seq 1 60); do \
		if curl -sf http://localhost:8080/realms/master > /dev/null 2>&1; then \
			echo "✓ Keycloak is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 60 ]; then \
			echo "❌ Keycloak failed to start"; \
			docker compose logs keycloak | tail -50; \
			exit 1; \
		fi; \
		echo "  Waiting... ($$i/60)"; \
		sleep 2; \
	done
	@echo "Configuring Keycloak..."
	./scripts/setup-keycloak.sh

clean-keycloak: ## Stop Keycloak and remove all data
	@echo "Stopping and removing Keycloak data..."
	docker compose down -v
	@rm -f .env
	@echo "Keycloak data removed"

# Build and dependency targets
deps: ## Download dependencies
	go mod download
	go mod verify

tidy: ## Tidy go.mod
	go mod tidy

# Cleanup targets
clean: ## Remove build artifacts and coverage files
	@rm -f coverage.out coverage-integration.out coverage.txt coverage.html
	@echo "Build artifacts cleaned"

# CI simulation
ci: fmt vet test-coverage ## Run CI checks locally
	@echo "✓ All CI checks passed"

# Development workflow
dev: ## Start development environment (Keycloak + setup)
	@echo "Setting up development environment..."
	@$(MAKE) setup-keycloak
	@echo ""
	@echo "✓ Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Load environment: source .env"
	@echo "  2. Run tests: make test-integration"
	@echo "  3. Access Keycloak: http://localhost:8080 (admin/admin)"

# Quick reference
.DEFAULT_GOAL := help
