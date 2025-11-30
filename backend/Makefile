.PHONY: dev build test lint fmt vet clean

# Development
dev:
	go run ./cmd/server/main.go

# Build
build:
	go build -o bin/server ./cmd/server/main.go

# Testing
test:
	go test -v ./...

test-short:
	go test -short ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting and formatting
lint: fmt vet

fmt:
	go fmt ./...

vet:
	go vet ./...

# Dependencies
tidy:
	go mod tidy

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migrations (requires golang-migrate)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

migrate-status:
	migrate -path db/migrations -database "$(DATABASE_URL)" version

# Help
help:
	@echo "Available targets:"
	@echo "  dev           - Run development server"
	@echo "  build         - Build binary"
	@echo "  test          - Run all tests"
	@echo "  test-short    - Run short tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run fmt and vet"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  tidy          - Tidy dependencies"
	@echo "  clean         - Clean build artifacts"
