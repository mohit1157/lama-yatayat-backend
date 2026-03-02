# ══════════════════════════════════════════════════
# LaMa Yatayat Backend - Makefile
# ══════════════════════════════════════════════════

.PHONY: help run build test migrate seed infra-up infra-down clean proto lint

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Infrastructure ─────────────────────────────
infra-up: ## Start Postgres + Redis (Docker)
	docker-compose up -d postgres redis
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 3
	@echo "✅ Infrastructure ready"

infra-down: ## Stop infrastructure
	docker-compose down

# ─── Database ───────────────────────────────────
migrate: ## Run database migrations
	@echo "🔄 Running migrations..."
	@for f in migrations/*.sql; do \
		echo "  → $$f"; \
		PGPASSWORD=devpassword123 psql -h localhost -U lamayatayat -d lamayatayat -f "$$f" 2>/dev/null || true; \
	done
	@echo "✅ Migrations complete"

seed: ## Seed demo data
	go run scripts/seed.go

# ─── Run Services ───────────────────────────────
run: ## Run all services locally (requires infra-up first)
	@echo "🚀 Starting all services..."
	@trap 'kill 0' EXIT; \
	go run ./cmd/user-service & \
	go run ./cmd/pricing-service & \
	go run ./cmd/geolocation-service & \
	go run ./cmd/ride-service & \
	go run ./cmd/route-matching & \
	go run ./cmd/payment-service & \
	go run ./cmd/notification-service & \
	go run ./cmd/connection-service & \
	wait

run-user: ## Run user service only
	go run ./cmd/user-service

run-ride: ## Run ride service only
	go run ./cmd/ride-service

run-matching: ## Run route matching service only
	go run ./cmd/route-matching

run-geo: ## Run geolocation service only
	go run ./cmd/geolocation-service

# ─── Build ──────────────────────────────────────
build: ## Build all services
	@mkdir -p bin
	@for svc in user-service ride-service route-matching geolocation-service payment-service pricing-service notification-service connection-service; do \
		echo "📦 Building $$svc..."; \
		go build -o bin/$$svc ./cmd/$$svc/; \
	done
	@echo "✅ All services built"

build-gateway: ## Build unified gateway (for deployment)
	@mkdir -p bin
	go build -o bin/gateway ./cmd/gateway/
	@echo "✅ Gateway built"

run-gateway: ## Run unified gateway locally
	go run ./cmd/gateway

# ─── Docker ─────────────────────────────────────
docker-up: ## Start everything with Docker Compose
	docker-compose up --build -d
	@echo "✅ All services running"

docker-down: ## Stop everything
	docker-compose down -v

docker-logs: ## Tail logs from all services
	docker-compose logs -f

# ─── Test ───────────────────────────────────────
test: ## Run all tests
	go test ./... -v -count=1

test-coverage: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# ─── Code Quality ───────────────────────────────
lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w .
	goimports -w .

# ─── Proto ──────────────────────────────────────
proto: ## Generate protobuf code
	@echo "🔄 Generating protobuf code..."
	buf generate
	@echo "✅ Protobuf code generated"

# ─── Clean ──────────────────────────────────────
clean: ## Clean build artifacts
	rm -rf bin/ tmp/ coverage.out coverage.html
