BIN_DIR ?= bin
BIN_FILE ?= chatmailer
COMPOSE_FILE ?= docker-compose.yaml
LINT_FILE ?= .golangci.yaml

.PHONY: up 
up: ## Deploy application in Docker via docker-compose configuration.
	docker compose -f $(COMPOSE_FILE) up --build -d

.PHONY: down
down: ## Bring down current Docker deployment.
	docker compose -f $(COMPOSE_FILE) down

.PHONY: run
run: ## Compile and run application binary.
	go run cmd/chatmailer/*

.PHONY: build
build: ## Compile application binary.
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_FILE) cmd/chatmailer/*

.PHONY: test
test: ## Compile and run application tests.
	go test -v ./...

.PHONY: format-lint format lint
format-lint: format lint ## Run both formatter and linter.
format: ## Run formatter only.
	gofumpt -l -w .
lint: ## Run linter only.
	golangci-lint run -c $(LINT_FILE) ./...

.PHONY: help
help: ## Display help information
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
