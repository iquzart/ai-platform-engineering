.DEFAULT_GOAL := help
.PHONY: help run test vet build docker-build compose-up compose-down

BINARY_NAME := chat-app

help: ## Display commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'

run: ## Run the service locally
	go run ./cmd/api

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

build: ## Build the service binary
	go build -o bin/$(BINARY_NAME) ./cmd/api

docker-build: ## Build the container image
	docker build -f Containerfile -t $(BINARY_NAME):local .

compose-up: ## Start services with Docker Compose
	docker compose -f docker-compose.yaml up --build

compose-down: ## Stop Docker Compose services
	docker compose -f docker-compose.yaml down
