.PHONY: build test test-unit test-integration test-coverage lint fmt fmt-check run run-local run-production compose-up compose-down tidy sqlc-generate sqlc-check migration-create migrate-up migrate-down migrate-status migrate-reset seed db-check

build:
	go build ./...

test:
	go test ./...

test-unit:
	go test ./...

test-integration:
	go test -tags=integration ./...

test-coverage:
	go test ./... -covermode=atomic -coverprofile=coverage.out

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

lint:
	go vet ./...

# Uses whatever is already in `.env` (and `internal/.env` if present).
run:
	go run ./cmd/api

# Copies `.env.local` over `.env`, then starts the API (overwrites `.env`).
run-local:
	@test -f .env.local || (echo "backend/.env.local not found; create it from .env.example" >&2; exit 1)
	cp -f .env.local .env
	go run ./cmd/api

# Copies `.env.prod` over `.env`, then starts the API (overwrites `.env`).
run-production:
	@test -f .env.prod || (echo "backend/.env.prod not found" >&2; exit 1)
	cp -f .env.prod .env
	go run ./cmd/api

sqlc-generate:
	sqlc generate

sqlc-check:
	sqlc generate
	git diff --exit-code -- internal/store sqlc.yaml

migration-create:
	@test -n "$(name)" || (echo "usage: make migration-create name=create_example_table" >&2; exit 1)
	goose -dir db/migrations create "$(name)" sql

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

migrate-reset:
	go run ./cmd/migrate reset

seed:
	go run ./cmd/seed

db-check:
	go run ./cmd/migrate up
	go run ./cmd/migrate status
	go run ./cmd/seed

compose-up:
	docker compose up -d

compose-down:
	docker compose down

tidy:
	go mod tidy
