.PHONY: install dev dev-backend dev-frontend test test-watch test-coverage lint typecheck build \
        docker-build docker-up docker-down db-init db-up db-down migrate migrate-down migrate-status \
        deploy deploy-frontend clean help

# Default environment
ENV ?= development

# Migrate binary (prefer go-installed version with sqlite3 support over homebrew)
MIGRATE := $(shell test -f "$(HOME)/go/bin/migrate" && echo "$(HOME)/go/bin/migrate" || which migrate 2>/dev/null)

# Database URL (default to SQLite for development)
# Use := to force evaluation, and provide default if empty
DB_URL := $(or $(DATABASE_URL),sqlite3://./data/conduit.db)

# ============================================================================
# Installation
# ============================================================================

install: install-backend install-frontend
	@echo "✅ All dependencies installed"

install-backend:
	@echo "📦 Installing backend dependencies..."
	cd backend && go mod download && go mod tidy

install-frontend:
	@echo "📦 Installing frontend dependencies..."
	cd frontend && npm install

# ============================================================================
# Development
# ============================================================================

dev:
	@./scripts/dev.sh

dev-backend:
	@echo "🚀 Starting backend server..."
	cd backend && go run ./cmd/server/main.go

dev-frontend:
	@echo "🚀 Starting frontend dev server..."
	cd frontend && npm run dev

# ============================================================================
# Testing
# ============================================================================

test: test-backend test-frontend
	@echo "✅ All tests passed"

test-backend:
	@echo "🧪 Running backend tests..."
	cd backend && go test -v ./...

test-frontend:
	@echo "🧪 Running frontend tests..."
	cd frontend && npm run test -- --run

test-watch:
	@echo "👀 Running frontend tests in watch mode..."
	cd frontend && npm run test

test-coverage:
	@echo "📊 Running tests with coverage..."
	cd backend && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	cd frontend && npm run test -- --run --coverage

# ============================================================================
# Linting & Type Checking
# ============================================================================

lint: lint-backend lint-frontend
	@echo "✅ Lint passed"

lint-backend:
	@echo "🔍 Linting backend..."
	cd backend && go fmt ./... && go vet ./...

lint-frontend:
	@echo "🔍 Linting frontend..."
	cd frontend && npm run lint

typecheck:
	@echo "🔍 Type checking frontend..."
	cd frontend && npm run typecheck

# ============================================================================
# Build
# ============================================================================

build: build-backend build-frontend
	@echo "✅ Build complete"

build-backend:
	@echo "🔨 Building backend..."
	cd backend && go build -o bin/server ./cmd/server/main.go

build-frontend:
	@echo "🔨 Building frontend..."
	cd frontend && npm run build

# ============================================================================
# Docker
# ============================================================================

docker-build:
	@echo "🐳 Building Docker images..."
	docker compose build

docker-up:
	@echo "🐳 Starting Docker containers..."
	docker compose up -d

docker-down:
	@echo "🐳 Stopping Docker containers..."
	docker compose down

# ============================================================================
# Database
# ============================================================================

db-init:
	@echo "🗄️ Initializing SQLite database..."
	mkdir -p backend/data
	touch backend/data/conduit.db
	@echo "✅ SQLite database initialized at backend/data/conduit.db"

db-up:
	@echo "🐘 Starting PostgreSQL container..."
	docker compose up -d postgres

db-down:
	@echo "🐘 Stopping PostgreSQL container..."
	docker compose down postgres

migrate:
	@echo "📦 Running migrations..."
	cd backend && $(MIGRATE) -path db/migrations -database "$(DB_URL)" up

migrate-down:
	@echo "📦 Rolling back last migration..."
	cd backend && $(MIGRATE) -path db/migrations -database "$(DB_URL)" down 1

migrate-status:
	@echo "📦 Migration status..."
	cd backend && $(MIGRATE) -path db/migrations -database "$(DB_URL)" version

migrate-create:
	@echo "📦 Creating new migration..."
	cd backend && $(MIGRATE) create -ext sql -dir db/migrations -seq $(NAME)

# ============================================================================
# Deployment
# ============================================================================

deploy:
	@echo "🚀 Deploying infrastructure..."
	cd infra && npx cdk deploy --all

deploy-frontend:
	@echo "🚀 Deploying frontend to GitHub Pages..."
	cd frontend && npm run build
	@echo "📤 Deploy frontend/dist to GitHub Pages"

# ============================================================================
# Cleanup
# ============================================================================

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf backend/bin backend/coverage.out backend/coverage.html
	rm -rf frontend/dist frontend/coverage
	rm -rf infra/cdk.out
	@echo "✅ Clean complete"

# ============================================================================
# Help
# ============================================================================

help:
	@echo "RealWorld Conduit - Available Commands"
	@echo ""
	@echo "Installation:"
	@echo "  make install          - Install all dependencies"
	@echo "  make install-backend  - Install backend dependencies"
	@echo "  make install-frontend - Install frontend dependencies"
	@echo ""
	@echo "Development:"
	@echo "  make dev              - Start both backend and frontend"
	@echo "  make dev-backend      - Start backend only"
	@echo "  make dev-frontend     - Start frontend only"
	@echo ""
	@echo "Testing:"
	@echo "  make test             - Run all tests"
	@echo "  make test-backend     - Run backend tests"
	@echo "  make test-frontend    - Run frontend tests"
	@echo "  make test-watch       - Run frontend tests in watch mode"
	@echo "  make test-coverage    - Run tests with coverage"
	@echo ""
	@echo "Quality:"
	@echo "  make lint             - Run linters"
	@echo "  make typecheck        - Run TypeScript type check"
	@echo ""
	@echo "Build:"
	@echo "  make build            - Build all"
	@echo "  make build-backend    - Build backend binary"
	@echo "  make build-frontend   - Build frontend bundle"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build     - Build Docker images"
	@echo "  make docker-up        - Start Docker containers"
	@echo "  make docker-down      - Stop Docker containers"
	@echo ""
	@echo "Database:"
	@echo "  make db-init          - Initialize SQLite database"
	@echo "  make db-up            - Start PostgreSQL container"
	@echo "  make db-down          - Stop PostgreSQL container"
	@echo "  make migrate          - Run database migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-status   - Show migration status"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy           - Deploy with AWS CDK"
	@echo "  make deploy-frontend  - Deploy frontend to GitHub Pages"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean            - Remove build artifacts"
