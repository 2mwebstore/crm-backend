# ─── CRM Backend Makefile ─────────────────────────────────────────────────────
APP      := crm-backend
BIN      := ./bin/$(APP)
MAIN     := ./cmd/server/main.go

.PHONY: run build tidy swag install-swag test lint env docker-up docker-down

## Start development server
run:
	@go run $(MAIN)

## Build production binary
build:
	@mkdir -p ./bin
	@go build -buildvcs=false -ldflags="-s -w" -o $(BIN) $(MAIN)
	@echo "✅  Built → $(BIN)"

## Download & tidy dependencies
tidy:
	@go mod tidy

## Generate Swagger docs
swag:
	@swag init -g $(MAIN) -o ./docs --parseDependency --parseInternal
	@echo "✅  Swagger docs → ./docs"

## Install swag CLI
install-swag:
	@go install github.com/swaggo/swag/cmd/swag@latest

## Run tests
test:
	@go test ./... -v -cover

## Lint (requires golangci-lint)
lint:
	@golangci-lint run ./...

## Copy .env.example → .env
env:
	@cp -n .env.example .env && echo "✅  .env created" || echo "⚠️   .env already exists"

## Start MySQL + phpMyAdmin via Docker Compose
docker-up:
	@docker compose up -d
	@echo "⏳  MySQL   → localhost:3306"
	@echo "🖥️   phpMyAdmin → http://localhost:8081"

## Stop Docker services
docker-down:
	@docker compose down

## Show all routes
routes:
	@go run $(MAIN) --routes 2>/dev/null || true
