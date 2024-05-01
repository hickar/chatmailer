.PHONY: up 
up:
	docker compose up --build

.PHONY: down
down:
	docker compose down

.PHONY: run
run:
	go run cmd/chatmailer

.PHONY: build
build:
	CGO_ENABLED=0 go build cmd/chatmailer

.PHONY: test
test:
	go test -v ./...

.PHONY: format-lint format lint
format-lint: format lint
format:
	gofumpt -l -w .
lint:
	golangci-lint run ./...
