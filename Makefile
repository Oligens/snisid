# SNISID — Master Makefile
.PHONY: all build build-all test test-short test-race coverage lint vet run-identity-api run-fraud-engine docker-up docker-down docker-logs docker-build db-migrate db-rollback clean deps deps-update help

all: lint test build

build: ## Compile les binaires Go réellement présents
	@mkdir -p bin
	go build -o bin/identity-api ./services/identity-api/cmd
	go build -o bin/fraud-engine ./services/fraud-engine/cmd
	go build -o bin/verification-api ./services/verification-api/cmd
	go build -o bin/audit-service ./services/audit-service/cmd
	go build -o bin/nexus-server ./services/nexus/cmd/nexus-server
	go build -o bin/ws-gateway ./services/ws-gateway

build-all: ## Compile tous les packages Go du workspace
	go build ./...

test: ## Exécute tous les tests Go
	go test ./... -count=1 -cover

test-short: ## Tests rapides
	go test ./... -count=1 -short -cover

test-race: ## Tests avec race detector
	go test ./... -count=1 -race -coverprofile=coverage.out

coverage: test-race
	go tool cover -html=coverage.out -o coverage.html

lint: ## Analyse statique sans ignorer les erreurs
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*' -not -path '*/node_modules/*')
	go vet ./...
	staticcheck ./...

vet:
	go vet ./...

run-identity-api:
	go run ./services/identity-api/cmd

run-fraud-engine:
	go run ./services/fraud-engine/cmd

docker-up:
	docker compose -f docker-compose.yml up -d

docker-down:
	docker compose -f docker-compose.yml down

docker-logs:
	docker compose -f docker-compose.yml logs -f

docker-build:
	docker compose -f docker-compose.yml build

db-migrate:
	go run ./scripts/migrations/main.go

db-rollback:
	go run ./scripts/migrations/main.go --down

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

deps:
	go mod download
	go mod tidy

deps-update:
	go get -u ./...
	go mod tidy

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' Makefile | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
